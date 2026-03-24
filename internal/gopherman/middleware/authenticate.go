// Package middleware provides reusable HTTP middleware.
package middleware

import (
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/gopherman/contextkey"
	"gophermart-loyalty/internal/gopherman/handler/api"
	"net/http"
)

// Authenticate validates bearer token and injects user ID into context.
func Authenticate(handler *api.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			if raw == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			userID, err := handler.UserIDFromRequest(r.Context(), raw)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if userID == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := contextkey.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
