package auth

import (
	"context"
	"errors"
	"figenn/internal/mailer"
	"figenn/internal/user"
	"log"
	"regexp"
	"time"

	"github.com/bluele/gcache"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	JWTSecret     string
	TokenDuration time.Duration
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

	if !isValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	if !isStrongPassword(req.Password) {
		return nil, ErrPasswordTooWeak
	}

	exists, err := s.repo.UserExistsByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	if exists {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	newUser := &user.User{
		Email:             req.Email,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Password:          string(hashedPassword),
		ProfilePictureUrl: "https://api.dicebear.com/7.x/initials/svg?seed=" + string(req.FirstName[0]) + string(req.LastName[0]),
		Country:           req.Country,
	}

	err = s.repo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	if cacheErr := s.cache.SetWithExpire(newUser.Email, newUser, time.Minute*5); cacheErr != nil {
		log.Println("Failed to cache user", cacheErr)
	}

	go s.sendWelcomeEmail(newUser)

	return &RegisterResponse{
		Message: "User created successfully",
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, ErrMissingFields
	}

	if !isValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.FirstName + " " + user.LastName,
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

func isValidEmail(email string) bool {
	return regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`).MatchString(email)
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)

	return hasNumber && hasUpper
}

func (s *Service) sendWelcomeEmail(user *user.User) {
	ctx := context.Background()

	emailConfig := mailer.Config{
		From:    "contact@alexandredissi.fr",
		To:      user.Email,
		Subject: "Bienvenue sur notre application",
		Html:    "<p>Bonjour " + user.FirstName + ",</p><p>Merci de vous être inscrit sur notre application.</p>",
	}

	_, err := s.mailer.SendMail(ctx, emailConfig)
	if err != nil {
		log.Println("Failed to send welcome email", err)
	}
}
