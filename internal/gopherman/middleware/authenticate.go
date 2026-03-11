package middleware

import (
	"gophermart-loyalty/internal/gopherman/handler/api"
	"net/http"
)

func Authenticate(handler *api.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			if raw == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ok, err := handler.ValidateSession(r.Context(), raw)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
