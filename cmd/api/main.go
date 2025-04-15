package main

import (
	"figenn/internal/database"
	"figenn/internal/server"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("La clé secrète JWT n'a pas été définie")
	}

	log.Println("Initialisation de la base de données...")
	db := database.New()
	defer db.Close()

	log.Println("Création du serveur...")
	config := server.Config{
		JWTSecret: jwtSecret,
	}
	srv := server.NewServer(db, config)
	srv.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Tentative de démarrage du serveur sur le port %s...", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}
