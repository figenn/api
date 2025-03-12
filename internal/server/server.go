package server

import (
	"figenn/internal/database"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// mockgen -source=internal/server/server.go -destination=internal/server/mocks/mock_server.go -package=mocks
type ServerStorer interface {
	Health() error
}

type Config struct {
	JWTSecret string
}

type Server struct {
	db     database.DbService
	router *echo.Echo
	config Config
}

func NewServer(db database.DbService, config Config) *Server {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return &Server{
		db:     db,
		router: e,
		config: config,
	}
}

func (s *Server) Start(port string) error {
	log.Printf("Server starting on port %s", port)
	return s.router.Start(":" + port)
}
