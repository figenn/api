// api.go (ton code de l'API user)
package users

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type API struct {
	JWTSecret string
	s         *Service
}

func NewAPI(secret string, service *Service) *API {
	return &API{
		JWTSecret: secret,
		s:         service,
	}
}

func (a *API) Bind(rg *echo.Group) {
	userGroup := rg.Group("/user", JWTMiddleware(a.JWTSecret))
	userGroup.GET("/me", a.Me)
}

func (a *API) Me(c echo.Context) error {
	ctx := c.Request().Context()
	userId := c.Get("user_id").(string)

	fmt.Println(userId)

	u, err := a.s.GetUserInfos(ctx, userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Utilisateur introuvable")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": u,
	})
}
