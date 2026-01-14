package middleware

import (
	"errors"
	"net/http"

	"github.com/0x0FACED/load-balancer/internal/limiter"
	"github.com/0x0FACED/load-balancer/internal/pkg/httpcommon"
)

type RateLimiterMiddleware struct {
	limiter limiter.RateLimitter
}

func NewRateLimiterMiddleware(limiter limiter.RateLimitter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimiterMiddleware) Limiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := httpcommon.ClientIDFromRequest(r)
		if !m.limiter.Allow(r.Context(), clientID) {
			httpcommon.JSONError(w, http.StatusTooManyRequests, errors.New("too many requests"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
