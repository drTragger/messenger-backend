package main

import (
	"github.com/drTragger/messenger-backend/internal/middleware"
	"github.com/drTragger/messenger-backend/internal/utils"
	"log"
	"net/http"
	"os"

	"github.com/drTragger/messenger-backend/config"
	"github.com/drTragger/messenger-backend/internal/handlers"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()

	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer db.Close()

	// Initialize translator
	translator := utils.NewTranslator()

	// Initialize repository and handler
	userRepo := repository.NewUserRepository(db)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET is not set in environment variables.")
	}

	authHandler := handlers.NewAuthHandler(userRepo, jwtSecret, translator)

	// Setup routes
	r := mux.NewRouter()
	r.Use(middleware.LanguageMiddleware(utils.FallbackLang))
	handlers.RegisterRoutes(r, authHandler)

	log.Printf("Server running on %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
