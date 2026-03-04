package middleware

import (
	"fmt"
	"net/http"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Authorization")
		if cookie == "" {
			fmt.Println("No Authorization header")
			return
		}
		fmt.Println("Authorization header:", cookie)
		next.ServeHTTP(w, r)
	})
}
