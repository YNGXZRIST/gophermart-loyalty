package middleware

import (
	"gophermart-loyalty/internal/gopherman/constant"
	"net/http"
)

func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := r.Header.Get(constant.ContentTypeHeader)
		if res != constant.ApplicationJSON {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		w.Header().Set(constant.ContentTypeHeader, constant.ApplicationJSON)
		next.ServeHTTP(w, r)

	})
}
