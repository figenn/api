package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type API struct {
	service *Service
}

func NewAPI(service *Service) *API {
	return &API{service: service}
}

func (a *API) Bind(rg *echo.Group) {
	authGroup := rg.Group("/auth")

	authGroup.POST("/register", a.Register)
	authGroup.POST("/login", a.Login)
	authGroup.POST("/forgot-password", a.ForgotPassword)
	authGroup.GET("/validate-reset-token", a.ValidateResetToken)
	authGroup.POST("/reset-password", a.ResetPassword)
	authGroup.GET("/logout", a.Logout)
	authGroup.POST("/refresh", a.RefreshToken)
}

func (a *API) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return jsonError(c, http.StatusBadRequest, ErrInvalidFormat)
	}

	if err := req.Validate(); err != nil {
		return jsonErrorStr(c, http.StatusUnprocessableEntity, err.Error())
	}

	resp, err := a.service.Register(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			return jsonError(c, http.StatusConflict, err)
		case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrPasswordTooWeak), errors.Is(err, ErrMissingFields):
			return jsonError(c, http.StatusUnprocessableEntity, err)
		default:
			return jsonError(c, http.StatusInternalServerError, ErrInternalServer)
		}
	}

	return c.JSON(http.StatusCreated, resp)
}

func (a *API) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return jsonErrorStr(c, http.StatusBadRequest, "Request format is invalid")
	}

	accessToken, refreshToken, err := a.service.Login(c.Request().Context(), req, c.Response().Writer)
	if err != nil {
		return jsonError(c, http.StatusUnauthorized, err)
	}

	setTokenCookies(c, *accessToken, *refreshToken, a.service.config)

	return c.JSON(http.StatusOK, echo.Map{"message": "Login successful"})
}

func (a *API) ForgotPassword(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return jsonErrorStr(c, http.StatusBadRequest, "Request format is invalid")
	}

	if req.Email == "" {
		return jsonError(c, http.StatusBadRequest, ErrMissingFields)
	}

	err := a.service.ForgotPassword(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return jsonError(c, http.StatusNotFound, err)
		default:
			return jsonError(c, http.StatusInternalServerError, ErrInternalServer)
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Password reset link has been sent to your email",
	})
}

func (a *API) ValidateResetToken(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusOK, echo.Map{"valid": false})
	}

	valid, err := a.service.IsValidResetToken(c.Request().Context(), token)
	if err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrInternalServer)
	}

	return c.JSON(http.StatusOK, echo.Map{"valid": valid})
}

func (a *API) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return jsonErrorStr(c, http.StatusBadRequest, "Request format is invalid")
	}

	if req.Token == "" || req.Password == "" {
		return jsonError(c, http.StatusBadRequest, ErrMissingFields)
	}

	err := a.service.ResetPassword(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenExpired), errors.Is(err, ErrInvalidToken):
			return jsonErrorStr(c, http.StatusUnauthorized, "Token is invalid or expired")
		default:
			return jsonError(c, http.StatusInternalServerError, ErrInternalServer)
		}
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Password has been reset"})
}

func (a *API) Logout(c echo.Context) error {
	a.service.Logout(c.Response())
	return c.NoContent(http.StatusOK)
}

func (a *API) RefreshToken(c echo.Context) error {
	refreshCookie, err := c.Cookie("refreshToken")
	if err != nil || refreshCookie.Value == "" {
		return jsonError(c, http.StatusUnauthorized, ErrInvalidToken)
	}

	accessToken, refreshToken, err := a.service.RefreshToken(c.Request().Context(), refreshCookie.Value)
	if err != nil {
		return jsonError(c, http.StatusUnauthorized, err)
	}

	setTokenCookies(c, *accessToken, *refreshToken, a.service.config)

	return c.JSON(http.StatusOK, echo.Map{"message": "Token refreshed"})
}

func setTokenCookies(c echo.Context, accessToken, refreshToken string, cfg Config) {
	secure := cfg.Environment == "production"

	c.SetCookie(&http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(cfg.TokenDuration),
	})

	c.SetCookie(&http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api",
		Expires:  time.Now().Add(cfg.RefreshTokenDuration),
	})
}

func jsonError(c echo.Context, code int, err error) error {
	return c.JSON(code, ErrorResponse{Error: err.Error()})
}

func jsonErrorStr(c echo.Context, code int, msg string) error {
	return c.JSON(code, ErrorResponse{Error: msg})
}
