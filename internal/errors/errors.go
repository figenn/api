package errors

import (
	"net/http"
	"strings"
)

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ApiError) Error() string {
	return e.Message
}

func NewNotFoundError(message string) *ApiError {
	if message == "" {
		message = "The requested resource was not found."
	}
	return NewApiError(http.StatusNotFound, message)
}

func NewBadRequestError(message string) *ApiError {
	if message == "" {
		message = "Invalid request data."
	}
	return NewApiError(http.StatusBadRequest, message)
}

func NewForbiddenError(message string) *ApiError {
	if message == "" {
		message = "You do not have permission to perform this action."
	}
	return NewApiError(http.StatusForbidden, message)
}

func NewUnauthorizedError(message string) *ApiError {
	if message == "" {
		message = "Authentication required or invalid token."
	}
	return NewApiError(http.StatusUnauthorized, message)
}

func NewInternalServerError(message string) *ApiError {
	if message == "" {
		message = "An unexpected error occurred."
	}
	return NewApiError(http.StatusInternalServerError, message)
}

func NewApiError(status int, message string) *ApiError {
	return &ApiError{
		Code:    status,
		Message: strings.TrimSpace(message),
	}
}
