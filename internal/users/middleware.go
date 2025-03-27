package users

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func CookieAuthMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessTokenCookie, err := c.Cookie("accessToken")
			if err != nil {
				fmt.Printf("Erreur lors de la récupération du cookie 'accessToken': %v\n", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
			}
			tokenString := accessTokenCookie.Value

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				if userID, ok := claims["user_id"].(string); ok {
					c.Set("user_id", userID)
				}
				if email, ok := claims["email"].(string); ok {
					c.Set("email", email)
				}
			}

			return next(c)
		}
	}
}
