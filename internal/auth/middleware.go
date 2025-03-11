package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTConfig contient la configuration du middleware JWT
type JWTConfig struct {
	Secret string
}

// JWT est un middleware qui vérifie les tokens JWT
func JWT(config JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Obtenir le header Authorization
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "Authentification requise",
				})
			}

			// Vérifier le format "Bearer token"
			parts := strings.SplitN(auth, " ", 2)
			if !(len(parts) == 2 && parts[0] == "Bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "Format d'authentification invalide",
				})
			}

			// Parser le token
			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
				}
				return []byte(config.Secret), nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "Token invalide: " + err.Error(),
				})
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Stocker les claims dans le contexte
				c.Set("user", claims)
				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Token invalide",
			})
		}
	}
}
