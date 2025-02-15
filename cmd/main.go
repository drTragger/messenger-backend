package main

import (
	"fmt"
	"github.com/drTragger/messenger-backend/internal/services"
	"github.com/drTragger/messenger-backend/internal/storage"
	"log"
	"net/http"
	"os"

	"github.com/drTragger/messenger-backend/config"
	"github.com/drTragger/messenger-backend/db"
	"github.com/drTragger/messenger-backend/internal/handlers"
	"github.com/drTragger/messenger-backend/internal/middleware"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/utils"
	"github.com/drTragger/messenger-backend/internal/websocket"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET is not set in environment variables.")
	}

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

	// Initialize websocket client manager
	clientManager := websocket.NewClientManager()

	// Initialize translator
	translator := utils.NewTranslator(getBasePath())

	// Initialize storage
	storageConfig := storage.Config{
		Type:      storage.LocalStorageType,
		LocalPath: storage.LocalStoragePath,
	}
	storageInst, err := storage.NewStorage(&storageConfig)
	if err != nil {
		log.Fatalf("Cannot initialize storage: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(pdb)
	tokenRepo := repository.NewTokenRepository(rdb)
	msgRepo := repository.NewMessageRepository(pdb)
	chatRepo := repository.NewChatRepository(pdb)
	attachmentRepo := repository.NewAttachmentRepository(pdb)

	// Initialize services
	msgService := services.NewMessageService(attachmentRepo, storageInst)
	wsService := services.NewWsService(clientManager)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, tokenRepo, jwtSecret, translator)
	messageHandler := handlers.NewMessageHandler(msgService, wsService, msgRepo, userRepo, chatRepo, attachmentRepo, storageInst, translator)
	chatHandler := handlers.NewChatHandler(chatRepo, userRepo, clientManager, translator)
	userHandler := handlers.NewUserHandler(userRepo, clientManager, storageInst, translator)
	wsHandler := handlers.NewWebSocketHandler(clientManager, tokenRepo, translator, jwtSecret)

	// Setup routes
	r := mux.NewRouter()
	r.Use(middleware.CORS())
	r.Use(middleware.LanguageMiddleware(utils.FallbackLang))
	handlers.RegisterRoutes(r, authHandler, messageHandler, chatHandler, userHandler, wsHandler)

	log.Printf("Server running on %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getBasePath() string {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	return cwd
}
