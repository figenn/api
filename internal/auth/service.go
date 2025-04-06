package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"figenn/internal/mailer"
	"figenn/internal/payment"
	"figenn/internal/users"
	"figenn/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/bluele/gcache"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *users.User) error
	CheckUserEmailExists(ctx context.Context, email string) (bool, error)
	FindUserByEmail(ctx context.Context, email string) (*users.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*users.User, error)
	InitDefaultSubscription(ctx context.Context, stripeCustomerID string) error
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string) error
	CheckRefreshToken(ctx context.Context, userID uuid.UUID, token string) (bool, error)
	SaveResetPasswordToken(ctx context.Context, userID uuid.UUID, token string) (uuid.UUID, string, error)
	IsResetTokenValid(ctx context.Context, token string) (bool, error)
	FindUserIDByResetToken(ctx context.Context, token string) (uuid.UUID, *string, bool, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, hashed string) error
	ClearResetToken(ctx context.Context, userID uuid.UUID) error
	StoreTOTPSecret(ctx context.Context, userID uuid.UUID, secret string) error
	RetrieveTOTPSecret(ctx context.Context, userID uuid.UUID) (string, error)
	EnableTOTP(ctx context.Context, userID uuid.UUID) error
	DisableTOTP(ctx context.Context, userID uuid.UUID) error
}

type Config struct {
	JWTSecret            string
	TokenDuration        time.Duration
	RefreshTokenDuration time.Duration
	AppURL               string
	Environment          string
}

type Service struct {
	repo   AuthRepository
	s      *payment.Service
	config Config
	cache  gcache.Cache
	mailer mailer.Mailer
}

func NewService(repo AuthRepository, config *Config, mailerClient mailer.Mailer, paymentService *payment.Service) *Service {
	return &Service{
		repo:   repo,
		config: *config,
		cache:  gcache.New(100).LRU().Expiration(time.Minute * 5).Build(),
		mailer: mailerClient,
		s:      paymentService,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	exists, err := s.repo.CheckUserEmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	stripeID, err := s.s.CreateCustomer(req.Email, req.FirstName, req.LastName)
	if err != nil {
		return nil, err
	}

	newUser := &users.User{
		Email:             req.Email,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Password:          hashedPassword,
		ProfilePictureUrl: "https://api.dicebear.com/7.x/initials/svg?seed=" + string(req.FirstName[0]) + string(req.LastName[0]),
		Country:           req.Country,
		StripeCustomerID:  *stripeID,
	}

	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}

	err = s.repo.InitDefaultSubscription(ctx, newUser.StripeCustomerID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.SetWithExpire(newUser.Email, newUser, 5*time.Minute)
	go utils.SendWelcomeEmail(s.mailer, newUser)

	return &RegisterResponse{Message: "User created successfully"}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest, w http.ResponseWriter) (*string, *string, error) {
	if req.Email == "" || req.Password == "" {
		return nil, nil, ErrMissingFields
	}
	if !utils.IsValidEmail(req.Email) {
		return nil, nil, ErrInvalidEmail
	}

	var user *users.User
	cachedUser, err := s.cache.Get(req.Email)
	if err == nil {
		if u, ok := cachedUser.(*users.User); ok {
			user = u
		}
	}
	if user == nil {
		userFromDB, err := s.repo.FindUserByEmail(ctx, req.Email)
		if err != nil {
			return nil, nil, err
		}
		_ = s.cache.SetWithExpire(req.Email, userFromDB, time.Minute*5)
		user = userFromDB
	}
	if !utils.ComparePassword(user.Password, req.Password) {
		return nil, nil, ErrInvalidCredentials
	}

	accessToken, err := generateToken(user, s.config.JWTSecret, s.config.TokenDuration)
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := generateRefreshToken(user, s.config.JWTSecret, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, nil, err
	}
	if err := s.repo.StoreRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, nil, ErrInternalServer
	}
	return &accessToken, &refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*string, *string, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, ErrInvalidToken
	}
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, nil, ErrInvalidToken
	}

	valid, err := s.repo.CheckRefreshToken(ctx, userID, refreshToken)
	if err != nil || !valid {
		return nil, nil, ErrInvalidToken
	}
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	newAccessToken, err := generateToken(user, s.config.JWTSecret, s.config.TokenDuration)
	if err != nil {
		return nil, nil, ErrInternalServer
	}
	newRefreshToken, err := generateRefreshToken(user, s.config.JWTSecret, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, nil, ErrInternalServer
	}
	if err := s.repo.StoreRefreshToken(ctx, user.ID, newRefreshToken); err != nil {
		return nil, nil, ErrInternalServer
	}
	return &newAccessToken, &newRefreshToken, nil
}

func (s *Service) Logout(w http.ResponseWriter) {
	expiration := time.Now().Add(-time.Hour)
	http.SetCookie(w, &http.Cookie{Name: "accessToken", Value: "", HttpOnly: true, Secure: s.config.Environment == "production", SameSite: http.SameSiteStrictMode, Path: "/", Expires: expiration})
	http.SetCookie(w, &http.Cookie{Name: "refreshToken", Value: "", HttpOnly: true, Secure: s.config.Environment == "production", SameSite: http.SameSiteStrictMode, Path: "/api/auth", Expires: expiration})
}

func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	if req.Email == "" || !utils.IsValidEmail(req.Email) {
		return ErrInvalidEmail
	}
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		fmt.Println("Error finding user:", err)
		return ErrUserNotFound
	}
	token := generateSecureToken()
	userID, tokenGenerated, err := s.repo.SaveResetPasswordToken(ctx, user.ID, token)
	if err != nil {
		fmt.Println("Error saving token:", err)
		return ErrInternalServer
	}
	err = s.cache.SetWithExpire(tokenGenerated, userID, 5*time.Minute)
	if err != nil {
		return ErrInternalServer
	}
	resetURL := s.config.AppURL + "/auth/reset-password?token=" + token
	go utils.SendResetPasswordEmail(s.mailer, user, resetURL)
	return nil
}

func (s *Service) IsValidResetToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}
	return s.repo.IsResetTokenValid(ctx, token)
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if req.Token == "" || req.Password == "" || !utils.IsStrongPassword(req.Password) {
		return ErrPasswordTooWeak
	}
	userID, email, valid, err := s.repo.FindUserIDByResetToken(ctx, req.Token)
	if err != nil || !valid {
		return ErrInvalidToken
	}
	_ = s.cache.Remove(*email)
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateUserPassword(ctx, userID, hashed); err != nil {
		return ErrInternalServer
	}
	_ = s.repo.ClearResetToken(ctx, userID)
	return nil
}

func generateToken(user *users.User, secret string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	})
	return token.SignedString([]byte(secret))
}

func generateRefreshToken(user *users.User, secret string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(duration).Unix(),
		"type":    "refresh",
		"iat":     time.Now().Unix(),
	})
	return token.SignedString([]byte(secret))
}

func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (s *Service) GenerateTOTPSecret(ctx context.Context, userID uuid.UUID) (string, string, error) {
	secret, err := totp.Generate(totp.GenerateOpts{Issuer: "Figenn", AccountName: "user" + userID.String()})
	if err != nil {
		return "", "", err
	}
	if err := s.repo.StoreTOTPSecret(ctx, userID, secret.Secret()); err != nil {
		return "", "", err
	}
	return secret.Secret(), secret.URL(), nil
}

func (s *Service) VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	secret, err := s.repo.RetrieveTOTPSecret(ctx, userID)
	if err != nil {
		return err
	}
	if !totp.Validate(code, secret) {
		return ErrInvalidTOTPCode
	}
	return nil
}

func (s *Service) EnableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	if err := s.VerifyTOTP(ctx, userID, code); err != nil {
		return err
	}
	return s.repo.EnableTOTP(ctx, userID)
}

func (s *Service) DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error {
	if err := s.VerifyTOTP(ctx, userID, code); err != nil {
		return err
	}
	return s.repo.DisableTOTP(ctx, userID)
}
