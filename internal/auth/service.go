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
	"log"
	"net/http"
	"time"

	"github.com/bluele/gcache"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type Config struct {
	JWTSecret            string
	TokenDuration        time.Duration
	RefreshTokenDuration time.Duration
	AppURL               string
	Environment          string
}

type Service struct {
	repo   *Repository
	s      *payment.Service
	config Config
	cache  gcache.Cache
	mailer mailer.Mailer
}

func NewService(repo *Repository, config *Config, mailerClient mailer.Mailer, paymentService *payment.Service) *Service {
	return &Service{
		repo:   repo,
		config: *config,
		cache:  gcache.New(100).LRU().Expiration(time.Minute * 5).Build(),
		mailer: mailerClient,
		s:      paymentService,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	exists, err := s.repo.UserExistsByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
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

	err = s.repo.CreateInitialSubscription(ctx, newUser.StripeCustomerID)
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
		userFromDB, err := s.repo.GetUserByEmail(ctx, req.Email)
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

	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, nil, ErrInternalServer
	}

	return &accessToken, &refreshToken, nil
}

func generateToken(user *users.User, secret string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateRefreshToken(user *users.User, secret string, duration time.Duration) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(duration).Unix(),
		"type":    "refresh", // Indicate it's a refresh token
		"iat":     time.Now().Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return refreshTokenString, nil
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

	valid, err := s.repo.VerifyRefreshToken(ctx, userID, refreshToken)
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

	if err := s.repo.SaveRefreshToken(ctx, user.ID, newRefreshToken); err != nil {
		return nil, nil, ErrInternalServer
	}

	return &newAccessToken, &newRefreshToken, nil
}

func (s *Service) Logout(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		HttpOnly: true,
		Secure:   s.config.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		HttpOnly: true,
		Secure:   s.config.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth",
		Expires:  time.Now().Add(-time.Hour),
	})
}

func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	if req.Email == "" {
		return ErrMissingFields
	}

	if !utils.IsValidEmail(req.Email) {
		return ErrInvalidEmail
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		return ErrInternalServer
	}

	token := generateSecureToken()
	idUser, tokenGenerated, err := s.repo.SavePasswordResetToken(ctx, user.ID, token)
	if err != nil {
		return ErrInternalServer
	}
	err = s.cache.SetWithExpire(tokenGenerated, idUser, time.Minute*5)
	if err != nil {
		return ErrInternalServer
	}

	resetUrl := s.config.AppURL + "/auth/reset-password?token=" + token
	go utils.SendResetPasswordEmail(s.mailer, user, resetUrl)

	return nil
}

func (s *Service) IsValidResetToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	return s.repo.IsValidResetToken(ctx, token)
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if req.Token == "" || req.Password == "" {
		return ErrMissingFields
	}

	if !utils.IsStrongPassword(req.Password) {
		return ErrPasswordTooWeak
	}

	var userID uuid.UUID
	id, email, valid, dbErr := s.repo.GetUserIDByResetToken(ctx, req.Token)
	if dbErr != nil || !valid {
		return ErrInvalidToken
	}

	if removed := s.cache.Remove(*email); !removed {
		log.Println("Failed to remove email from cache")
	}
	userID = uuid.UUID(id)

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	err = s.repo.ResetPassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return ErrInternalServer
	}

	err = s.repo.ClearResetToken(ctx, userID)
	if err != nil {
		log.Println("Failed to clear reset token:", err)
	}

	return nil
}

func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
