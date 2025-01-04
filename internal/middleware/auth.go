package middleware

import (
	"context"
	"github.com/drTragger/messenger-backend/internal/responses"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := strings.TrimSpace(strings.Replace(r.Header.Get("Authorization"), "Bearer", "", 1))
			if tokenString == "" {
				responses.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Token not provided")
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				responses.ErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Invalid token")
				return
			}

			claims := token.Claims.(jwt.MapClaims)
			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
