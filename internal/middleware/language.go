package middleware

import (
	"context"
	"net/http"
)

type ContextKey string

const LanguageKey ContextKey = "language"

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

			ctx := context.WithValue(r.Context(), LanguageKey, lang)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
