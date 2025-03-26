package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct {
	service *Service
}

func NewAPI(service *Service) *API {
	return &API{
		service: service,
	}
}

func (a *API) Bind(rg *echo.Group) {
	authGroup := rg.Group("/auth")
	authGroup.POST("/register", a.Register)
	authGroup.POST("/login", a.Login)
	authGroup.POST("/forgot-password", a.ForgotPassword)
	authGroup.GET("/validate-reset-token", a.ValidateResetToken)
	authGroup.POST("/reset-password", a.ResetPassword)
}

func (a *API) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Request format is invalid",
		})
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Request format is invalid",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrMissingFields.Error(),
		})
	}

	resp, err := a.service.Register(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error: err.Error(),
			})

		case errors.Is(err, ErrInvalidEmail) || errors.Is(err, ErrPasswordTooWeak) || errors.Is(err, ErrMissingFields):
			return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Error: err.Error(),
			})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrInternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusCreated, resp)
}

func (a *API) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Request format is invalid",
		})
	}

	resp, err := a.service.Login(c.Request().Context(), req, c.Response().Writer)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (a *API) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Request format is invalid",
		})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrMissingFields.Error(),
		})
	}

	err := a.service.ForgotPassword(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: err.Error(),
			})

		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrInternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Password reset link has been sent to your email",
	})
}

func (a *API) ValidateResetToken(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Token is missing",
		})
	}

	err := a.service.ValidateResetToken(c.Request().Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrInvalidToken):
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: err.Error(),
			})

		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrInternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Token is valid",
	})
}

func (a *API) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Request format is invalid",
		})
	}

	if req.Token == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrMissingFields.Error(),
		})
	}

	err := a.service.ResetPassword(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenExpired) || errors.Is(err, ErrInvalidToken):
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: "Token is invalid or expired",
			})

		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: ErrInternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Password has been reset",
	})
}
