package server

import (
	"figenn/internal/auth"
	"figenn/internal/mailer"
	"figenn/internal/stripe"
	"figenn/internal/users"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) SetupRoutes() {
	apiGroup := s.router.Group("/api")

	apiGroup.GET("/health", s.healthHandler)

	s.setupAuthRoutes(apiGroup)

	s.setupUserRoutes(apiGroup)
	s.setupStripeRoutes(apiGroup)
}

func (s *Server) setupAuthRoutes(apiGroup *echo.Group) {
	authAPI := s.newAuthAPI()
	authAPI.Bind(apiGroup)
}

func (s *Server) setupUserRoutes(apiGroup *echo.Group) {
	userAPI := s.newUserAPI()
	userAPI.Bind(apiGroup)
}

func (s *Server) setupStripeRoutes(apiGroup *echo.Group) {
	stripeAPI := s.newStripeAPI()
	stripeAPI.Bind(apiGroup)
}

func (s *Server) newAuthAPI() *auth.API {
	authRepo := auth.NewRepository(s.db)
	authService := auth.NewService(authRepo, &auth.Config{
		JWTSecret:     s.config.JWTSecret,
		TokenDuration: time.Hour * 24 * 5, // 5 jours
		AppURL:        os.Getenv("APP_URL"),
	}, mailer.NewMailer(os.Getenv("RESEND_API_KEY")))

	return auth.NewAPI(authService)
}

func (s *Server) newUserAPI() *users.API {
	userRepo := users.NewRepository(s.db)
	authService := users.NewService(userRepo)
	return users.NewAPI(s.config.JWTSecret, authService)
}

func (s *Server) newStripeAPI() *stripe.API {
	stripeService := stripe.NewStripeClient(os.Getenv("STRIPE_SECRET_KEY"))
	return stripe.NewAPI(stripeService)
}

func (s *Server) healthHandler(c echo.Context) error {
	fmt.Println("Health check")
	return c.JSON(http.StatusOK, s.db.Health())
}
