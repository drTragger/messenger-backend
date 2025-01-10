package handlers

import (
	"github.com/drTragger/messenger-backend/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterRoutes(r *mux.Router, authHandler *AuthHandler, messageHandler *MessageHandler, wsHandler *WebSocketHandler) {
	apiRouter := r.PathPrefix("/api").Subrouter()
	authApiRouter := apiRouter.PathPrefix("/").Subrouter()
	authApiRouter.Use(middleware.Auth(authHandler.Secret, authHandler.TokenRepo, authHandler.Trans))

	// Auth routes
	apiRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/refresh-token", authHandler.RefreshToken).Methods("POST", "OPTIONS")
	authApiRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/phone/verify", authHandler.VerifyCode).Methods("POST", "OPTIONS")

	// Message routes
	authApiRouter.HandleFunc("/messages", messageHandler.SendMessage).Methods("POST", "OPTIONS")
	authApiRouter.HandleFunc("/messages", messageHandler.GetMessages).Methods("GET", "OPTIONS")

	// WebSocket routes
	r.Handle(
		"/ws",
		middleware.Auth(
			authHandler.Secret,
			authHandler.TokenRepo,
			authHandler.Trans,
		)(http.HandlerFunc(wsHandler.HandleWebSocket)),
	).Methods("GET")
}
