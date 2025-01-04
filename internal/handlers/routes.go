package handlers

import (
	"github.com/drTragger/messenger-backend/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterRoutes(r *mux.Router, authHandler *AuthHandler, messageHandler *MessageHandler, wsHandler *WebSocketHandler) {
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Auth routes
	apiRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	apiRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	apiRouter.HandleFunc("/refresh-token", authHandler.RefreshToken).Methods("POST")
	apiRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	apiRouter.HandleFunc("/phone/verify", authHandler.VerifyCode).Methods("POST")

	// Message routes
	apiRouter.Handle("/messages",
		middleware.Auth(authHandler.Secret, authHandler.TokenRepo, authHandler.Trans)(
			http.HandlerFunc(messageHandler.SendMessage),
		),
	).Methods("POST")

	// WebSocket routes
	r.Handle("/ws", middleware.Auth(authHandler.Secret, authHandler.TokenRepo, authHandler.Trans)(http.HandlerFunc(wsHandler.HandleWebSocket))).Methods("GET")
}
