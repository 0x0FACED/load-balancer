package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/0x0FACED/load-balancer/internal/limitter"
)

type RateLimiterMiddleware struct {
	limiter limitter.RateLimitter
}

func NewRateLimiterMiddleware(limiter limitter.RateLimitter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimiterMiddleware) Limitter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := clientIDFromRequest(r)
		fmt.Println("Client ID:", clientID) // test log
		if !m.limiter.Allow(clientID) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIDFromRequest(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}
