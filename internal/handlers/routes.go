package handlers

import (
	"github.com/drTragger/messenger-backend/internal/middleware"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, authHandler *AuthHandler, messageHandler *MessageHandler, chatHandler *ChatHandler, userHandler *UserHandler, wsHandler *WebSocketHandler) {
	apiRouter := r.PathPrefix("/api").Subrouter()
	authApiRouter := apiRouter.PathPrefix("/").Subrouter()
	authApiRouter.Use(middleware.Auth(authHandler.Secret, authHandler.TokenRepo, authHandler.UserRepo, authHandler.Trans))

	// Auth routes
	apiRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/refresh-token", authHandler.RefreshToken).Methods("POST", "OPTIONS")
	authApiRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/phone/verify", authHandler.VerifyCode).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/phone/verify/resend", authHandler.ResendCode).Methods("POST", "OPTIONS")
	authApiRouter.HandleFunc("/auth/me", authHandler.GetCurrentUser).Methods("GET", "OPTIONS")

	// Chat routes
	authApiRouter.HandleFunc("/chats", chatHandler.Create).Methods("POST", "OPTIONS")
	authApiRouter.HandleFunc("/chats", chatHandler.GetForUser).Methods("GET", "OPTIONS")
	authApiRouter.HandleFunc("/chats/{id}", chatHandler.GetByID).Methods("GET", "OPTIONS")

	// Message routes
	authApiRouter.HandleFunc("/chats/{chatId}/messages", messageHandler.SendMessage).Methods("POST", "OPTIONS")
	authApiRouter.HandleFunc("/chats/{chatId}/messages", messageHandler.GetMessages).Methods("GET", "OPTIONS")
	authApiRouter.HandleFunc("/chats/{chatId}/messages/{messageId}", messageHandler.EditMessage).Methods("PATCH", "OPTIONS")
	authApiRouter.HandleFunc("/chats/{chatId}/messages/{messageId}", messageHandler.DeleteMessage).Methods("DELETE", "OPTIONS")
	authApiRouter.HandleFunc("/chats/{chatId}/messages/{messageId}/read", messageHandler.MarkMessageRead).Methods("PATCH", "OPTIONS")

	// User routes
	authApiRouter.HandleFunc("/users", userHandler.GetUsers).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/users/profile-picture/{filename}", userHandler.GetProfilePicture).Methods("GET", "OPTIONS")
	authApiRouter.HandleFunc("/users/profile-picture", userHandler.UpdateProfilePicture).Methods("PATCH", "OPTIONS")
	authApiRouter.HandleFunc("/users/profile-picture", userHandler.DeleteProfilePicture).Methods("DELETE", "OPTIONS")

	// WebSocket routes
	r.HandleFunc("/ws", wsHandler.HandleWebSocket).Methods("GET")
	authApiRouter.HandleFunc("/users/online", wsHandler.GetOnlineUsers).Methods("GET", "OPTIONS")
	authApiRouter.HandleFunc("/users/online/{id}", wsHandler.GetUserIsOnline).Methods("GET", "OPTIONS")
}
