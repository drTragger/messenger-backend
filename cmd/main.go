package main

import (
	"fmt"
	"github.com/drTragger/messenger-backend/db"
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

	// Initialize Postgres DB
	pdb, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer pdb.Close()

	// Initialize Redis DB
	rdb, err := db.InitRedis(fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort))
	if err != nil {
		log.Fatalf("Cannot connect to Redis: %v", err)
	}

	// Initialize translator
	translator := utils.NewTranslator()

	// Initialize repository and handler
	userRepo := repository.NewUserRepository(pdb)
	tokenRepo := repository.NewTokenRepository(rdb)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET is not set in environment variables.")
	}

	authHandler := handlers.NewAuthHandler(userRepo, tokenRepo, jwtSecret, translator)

	// Setup routes
	r := mux.NewRouter()
	r.Use(middleware.LanguageMiddleware(utils.FallbackLang))
	handlers.RegisterRoutes(r, authHandler)

	log.Printf("Server running on %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
