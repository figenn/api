package auth

import (
	"errors"
	"figenn/internal/user"
	"fmt"
	"regexp"
	"time"

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
}

func NewService(repo *Repository, config Config) *Service {
	return &Service{
		repo:   repo,
		config: config,
	}
}

func (s *Service) Register(req RegisterRequest) (*RegisterResponse, error) {
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return nil, ErrMissingFields
	}

	if !isValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	if !isStrongPassword(req.Password) {
		return nil, ErrPasswordTooWeak
	}

	exists, err := s.repo.UserExistsByEmail(req.Email)
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
	}

	if exists {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, err
	}

	newUser := &user.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  string(hashedPassword),
	}

	err = s.repo.CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{
		Message: "User created successfully",
	}, nil
}

func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, ErrMissingFields
	}

	if !isValidEmail(req.Email) {
		return nil, ErrInvalidEmail
	}

	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
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
