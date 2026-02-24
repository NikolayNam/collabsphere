package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(5, 10)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
