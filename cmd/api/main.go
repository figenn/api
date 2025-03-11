package main

import (
	"figenn/internal/database"
	"figenn/internal/server"
	"log"
	"os"
)

func main() {
	log.Println("Application démarrage...")

	// Récupération des variables d'environnement
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_jwt_secret_for_development"
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET env var for production.")
	}

	log.Println("Initialisation de la base de données...")
	// Initialisation de la base de données
	db := database.New()
	defer db.Close()

	log.Println("Création du serveur...")
	// Initialisation du serveur avec la configuration
	config := server.Config{
		JWTSecret: jwtSecret,
	}
	srv := server.NewServer(db, config)

	log.Println("Configuration des routes...")
	// Configuration des routes
	srv.SetupRoutes()

	// Démarrage du serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Tentative de démarrage du serveur sur le port %s...", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}
