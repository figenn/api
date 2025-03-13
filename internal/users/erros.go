package users

import "errors"

var (
	ErrNoRows = errors.New("no rows found")
	ErrInternalServer     = errors.New("internal server error")
	ErrMissingFields      = errors.New("required fields missing")
	ErrDatabaseOperation  = errors.New("database operation failed")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNotFound           = errors.New("not found")
)
