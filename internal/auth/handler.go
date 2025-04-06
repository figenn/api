package auth

import (
	"errors"
	"figenn/internal/users"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type API struct {
	service   *Service
	JWTSecret string
}

func NewAPI(service *Service, secret string) *API {
	return &API{
		service:   service,
		JWTSecret: secret,
	}
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
	authGroup.POST("/enable-totp", a.EnableTOTP, users.CookieAuthMiddleware(a.service.config.JWTSecret))
	authGroup.POST("/disable-totp", a.DisableTOTP, users.CookieAuthMiddleware(a.service.config.JWTSecret))
	authGroup.POST("/verify-totp", a.VerifyTOTP, users.CookieAuthMiddleware(a.service.config.JWTSecret))
}

func (a *API) Register(c echo.Context) error {
	ctx := c.Request().Context()
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": ErrInvalidFormat.Error()})
	}
	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
	}
	resp, err := a.service.Register(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			return c.JSON(http.StatusConflict, echo.Map{"error": err.Error()})
		case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrPasswordTooWeak), errors.Is(err, ErrMissingFields):
			return c.JSON(http.StatusUnprocessableEntity, echo.Map{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": ErrInternalServer.Error()})
		}
	}
	return c.JSON(http.StatusCreated, resp)
}

func (a *API) Login(c echo.Context) error {
	ctx := c.Request().Context()
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Request format is invalid"})
	}
	accessToken, refreshToken, err := a.service.Login(ctx, req, c.Response().Writer)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}
	setTokenCookies(c, *accessToken, *refreshToken, a.service.config)
	return c.JSON(http.StatusOK, echo.Map{"message": "Login successful"})
}

func (a *API) ForgotPassword(c echo.Context) error {
	ctx := c.Request().Context()
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Request format is invalid"})
	}
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": ErrMissingFields.Error()})
	}
	err := a.service.ForgotPassword(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": ErrInternalServer.Error()})
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Password reset link has been sent to your email"})
}

func (a *API) ValidateResetToken(c echo.Context) error {
	ctx := c.Request().Context()
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusOK, echo.Map{"valid": false})
	}
	valid, err := a.service.IsValidResetToken(ctx, token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": ErrInternalServer.Error()})
	}

	if !valid {
		return c.JSON(http.StatusUnauthorized, echo.Map{"valid": false})
	}

	return c.JSON(http.StatusOK, echo.Map{"valid": valid})
}

func (a *API) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Request format is invalid"})
	}
	if req.Token == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": ErrMissingFields.Error()})
	}
	err := a.service.ResetPassword(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenExpired), errors.Is(err, ErrInvalidToken):
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Token is invalid or expired"})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": ErrInternalServer.Error()})
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Password has been reset"})
}

func (a *API) Logout(c echo.Context) error {
	a.service.Logout(c.Response())
	return c.NoContent(http.StatusOK)
}

func (a *API) RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()
	refreshCookie, err := c.Cookie("refreshToken")
	if err != nil || refreshCookie.Value == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": ErrInvalidToken.Error()})
	}
	accessToken, refreshToken, err := a.service.RefreshToken(ctx, refreshCookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}
	setTokenCookies(c, *accessToken, *refreshToken, a.service.config)
	return c.JSON(http.StatusOK, echo.Map{"message": "Token refreshed"})
}

func (a *API) EnableTOTP(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := c.Get("user_id").(string)
	if !ok || userIDStr == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing user_id in context"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid user_id"})
	}
	secret, qr, err := a.service.GenerateTOTPSecret(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, TOTPSecretResponse{Secret: secret, QR: qr})
}

func (a *API) VerifyTOTP(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := c.Get("user_id").(string)
	if !ok || userIDStr == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing user_id in context"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid user_id"})
	}
	var req TOTPRequest
	if err := c.Bind(&req); err != nil || req.Code == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid TOTP payload"})
	}
	if err := a.service.EnableTOTP(ctx, userID, req.Code); err != nil {
		if errors.Is(err, ErrInvalidTOTPCode) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "TOTP enabled"})
}

func (a *API) DisableTOTP(c echo.Context) error {
	ctx := c.Request().Context()
	userIDStr, ok := c.Get("user_id").(string)
	if !ok || userIDStr == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Missing user_id in context"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid user_id"})
	}
	var req TOTPRequest
	if err := c.Bind(&req); err != nil || req.Code == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid TOTP payload"})
	}
	if err := a.service.DisableTOTP(ctx, userID, req.Code); err != nil {
		if errors.Is(err, ErrInvalidTOTPCode) {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.NoContent(http.StatusOK)
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
