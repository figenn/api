package users

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func CookieAuthMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("accessToken")
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing access token")
			}

			token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims")
			}

			userID, ok := claims["user_id"].(string)
			if !ok || userID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing user ID in token")
			}

			c.Set("user_id", userID)

			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}

			return next(c)
		}
	}
}

func PremiumMiddleware(s *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Get("user_id")
			if userID == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			user, err := s.repo.GetUser(c.Request().Context(), userID.(string))
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
			}

			ok, err := s.IsPremiumUser(user.StripeCustomerID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Unable to check subscription")
			}
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "Access restricted to premium users")
			}

			return next(c)
		}
	}
}
