package handlers

import (
	"github.com/drTragger/messenger-backend/internal/middleware"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterRoutes(r *mux.Router, authHandler *AuthHandler) {
	apiRouter := r.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	apiRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	apiRouter.HandleFunc("/refresh-token", authHandler.RefreshToken).Methods("POST")
	apiRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	apiRouter.HandleFunc("/phone/verify", authHandler.VerifyCode).Methods("POST")

	// Example of a protected route
	apiRouter.Handle("/profile", middleware.Auth(authHandler.Secret, authHandler.TokenRepo, authHandler.Trans)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		responses.SuccessResponse(w, http.StatusOK, "Profile fetched successfully", map[string]interface{}{"userId": userID})
	}))).Methods("GET")
}
