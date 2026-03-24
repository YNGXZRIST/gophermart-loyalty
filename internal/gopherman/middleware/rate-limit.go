package middleware

import (
	"net/http"
	"strconv"

	"golang.org/x/time/rate"
)

// RequestLimit limits request rate and returns 429 with Retry-After header.
func RequestLimit(limiter *rate.Limiter, retryAfter int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
