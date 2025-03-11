package server

import (
	"figenn/internal/auth"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) SetupRoutes() {
	apiGroup := s.router.Group("/api")

	s.router.GET("/health", s.healthHandler)

	authRepo := auth.NewRepository(s.db)
	authService := auth.NewService(authRepo, auth.Config{
		JWTSecret:     s.config.JWTSecret,
		TokenDuration: time.Hour * 24 * 7, // 7 jours
	})

	authAPI := auth.NewAPI(authService)
	authAPI.Bind(apiGroup)

}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
