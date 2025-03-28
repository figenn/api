package auth

import "strings"

type RegisterRequest struct {
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
	Email     string `json:"email" form:"email"`
	Password  string `json:"password" form:"password"`
	Country   string `json:"country" form:"country"`
}

func (r RegisterRequest) Validate() error {
	if r.Email == "" || r.Password == "" {
		return ErrMissingFields
	}

	if !strings.Contains(r.Email, "@") {
		return ErrInvalidEmail
	}

	if len(r.Password) < 8 {
		return ErrPasswordTooWeak
	}

	return nil
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type LoginRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" form:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" form:"token"`
	Password string `json:"password" form:"password"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}
