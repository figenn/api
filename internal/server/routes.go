package server

import (
	"figenn/internal/auth"
	"figenn/internal/mailer"
	"figenn/internal/powens"
	"figenn/internal/stripe"
	"figenn/internal/subscriptions"
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
	s.SetupPowensApi().Bind(apiGroup)
	s.setupSubscriptionRoutes(apiGroup)

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

func (s *Server) setupSubscriptionRoutes(apiGroup *echo.Group) {
	subscriptionAPI := s.SetupSubscriptionAPI()
	subscriptionAPI.Bind(apiGroup)
}

func (s *Server) newAuthAPI() *auth.API {
	authRepo := auth.NewRepository(s.db)
	stripeService := stripe.NewService(os.Getenv("STRIPE_SECRET_KEY"), users.NewRepository(s.db))
	authService := auth.NewService(authRepo, &auth.Config{
		JWTSecret:     s.config.JWTSecret,
		TokenDuration: time.Hour * 24 * 7, // 5 jours
		AppURL:        os.Getenv("APP_URL"),
	}, mailer.NewMailer(), stripeService)

	return auth.NewAPI(authService)
}

func (s *Server) newUserAPI() *users.API {
	userRepo := users.NewRepository(s.db)
	authService := users.NewService(userRepo)
	return users.NewAPI(s.config.JWTSecret, authService)
}

func (s *Server) newStripeAPI() *stripe.API {
	userRepo := users.NewRepository(s.db)
	stripeService := stripe.NewService(os.Getenv("STRIPE_SECRET_KEY"), userRepo)
	return stripe.NewAPI(s.config.JWTSecret, stripeService)
}

func (s *Server) SetupPowensApi() *powens.API {
	clientID := os.Getenv("POWENS_CLIENT_ID")
	clientSecret := os.Getenv("POWENS_CLIENT_SECRET")
	domain := os.Getenv("POWENS_DOMAIN")
	callbackURI := os.Getenv("POWENS_REDIRECT_URI")

	config := &powens.Config{
		ClientID:    clientID,
		Domain:      domain,
		CallbackURI: callbackURI,
	}

	client := powens.NewClient(clientID, clientSecret, http.DefaultClient)

	repo := powens.NewRepository(s.db)

	subscriptionsRepo := subscriptions.NewRepository(s.db)
	service := powens.NewService(repo, client, config, subscriptionsRepo)

	return powens.NewAPI(s.config.JWTSecret, service)
}

func (s *Server) SetupSubscriptionAPI() *subscriptions.API {
	subscriptionsRepo := subscriptions.NewRepository(s.db)
	subscriptionsService := subscriptions.NewService(subscriptionsRepo)
	return subscriptions.NewAPI(s.config.JWTSecret, subscriptionsService)
}

func (s *Server) healthHandler(c echo.Context) error {
	fmt.Println("Health check")
	return c.JSON(http.StatusOK, s.db.Health())
}
