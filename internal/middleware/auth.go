package middleware

import (
	"context"
	"github.com/drTragger/messenger-backend/internal/repository"
	"github.com/drTragger/messenger-backend/internal/responses"
	"github.com/drTragger/messenger-backend/internal/utils"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func Auth(secret string, tokenRepo *repository.TokenRepository, trans *utils.Translator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
			if tokenString == "" {
				responses.ErrorResponse(w, http.StatusUnauthorized, trans.Translate(r, "errors.unauthorized", nil), "Token not provided")
				return
			}

			// Parse the token
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				responses.ErrorResponse(w, http.StatusUnauthorized, trans.Translate(r, "errors.unauthorized", nil), "Invalid token")
				return
			}

			claims := token.Claims.(jwt.MapClaims)
			userID := uint(claims["user_id"].(float64))

			// Verify the token in Redis
			valid, err := tokenRepo.IsTokenValid(r.Context(), tokenString, userID)
			if err != nil || !valid {
				responses.ErrorResponse(w, http.StatusUnauthorized, trans.Translate(r, "errors.unauthorized", nil), "Token is invalid or expired")
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
