package middleware

import (
	"context"
	"github.com/drTragger/messenger-backend/internal/utils"
	"net/http"
)

func LanguageMiddleware(defaultLang string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := r.Header.Get("Accept-Language")
			if lang == "" {
				lang = r.URL.Query().Get("lang")
			}
			if lang == "" {
				lang = defaultLang
			}

			ctx := context.WithValue(r.Context(), utils.LanguageKey, lang)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
