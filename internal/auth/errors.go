package auth

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPasswordTooWeak    = errors.New("password too weak (minimum 8 characters, one uppercase letter, one number)")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrUserNotFound       = errors.New("user not found")
	ErrInternalServer     = errors.New("internal server error")
	ErrMissingFields      = errors.New("required fields missing")
	ErrDatabaseOperation  = errors.New("database operation failed")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
)
