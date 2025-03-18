package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"figenn/internal/mailer"
	"figenn/internal/users"
	"figenn/internal/utils"
	"log"
	"time"

	"github.com/bluele/gcache"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	JWTSecret     string
	TokenDuration time.Duration
	AppURL        string
}

type Service struct {
	repo   *Repository
	config Config
	cache  gcache.Cache
	mailer mailer.Mailer
}

func NewService(repo *Repository, config *Config, mailerClient mailer.Mailer) *Service {
	return &Service{
		repo:   repo,
		config: *config,
		cache:  gcache.New(100).LRU().Expiration(time.Minute * 5).Build(),
		mailer: mailerClient,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return nil, ErrMissingFields
	}

	if !utils.IsValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	if !utils.IsStrongPassword(req.Password) {
		return nil, ErrPasswordTooWeak
	}

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

	newUser := &users.User{
		Email:             req.Email,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Password:          hashedPassword,
		ProfilePictureUrl: "https://api.dicebear.com/7.x/initials/svg?seed=" + string(req.FirstName[0]) + string(req.LastName[0]),
		Country:           req.Country,
		Subscription:      users.Free,
	}

	err = s.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	if cacheErr := s.cache.SetWithExpire(newUser.Email, newUser, time.Minute*5); cacheErr != nil {
		log.Println("Failed to cache user", cacheErr)
	}

	go utils.SendWelcomeEmail(s.mailer, newUser)

	return &RegisterResponse{
		Message: "User created successfully",
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, ErrMissingFields
	}

	if !utils.IsValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !utils.ComparePassword(user.Password, req.Password) {
		return nil, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(s.config.TokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: tokenString,
	}, nil
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
		log.Println("Failed to cache reset token", err)
	}

	resetUrl := s.config.AppURL + "/auth/reset-password?token=" + token
	go utils.SendResetPasswordEmail(s.mailer, user, resetUrl)

	return nil
}

func (s *Service) ValidateResetToken(ctx context.Context, token string) error {
	if token == "" {
		return ErrMissingFields
	}

	tokenValue, err := s.cache.Get(token)
	if err == nil {
		_, ok := tokenValue.(int)
		if ok {
			return nil
		}
	}

	err = s.repo.ValidateResetToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if req.Token == "" || req.Password == "" {
		return ErrMissingFields
	}

	if !utils.IsStrongPassword(req.Password) {
		return ErrPasswordTooWeak
	}

	var userID uuid.UUID
	tokenValue, err := s.cache.Get(req.Token)
	if err == nil {
		id, ok := tokenValue.(uuid.UUID)
		if ok {
			userID = id
		}
	}

	if userID == uuid.Nil {
		id, valid, dbErr := s.repo.GetUserIDByResetToken(ctx, req.Token)
		if dbErr != nil || !valid {
			return ErrInvalidToken
		}
		userID = uuid.UUID(id)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
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
