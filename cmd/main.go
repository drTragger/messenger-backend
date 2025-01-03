package main

import (
	"log"
	"net/http"

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

	r := mux.NewRouter()

	handlers.RegisterRoutes(r)

	log.Printf("Server running on %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
