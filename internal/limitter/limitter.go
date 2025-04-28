package limitter

import "context"

type RateLimitter interface {
	Allow(clientID string) bool
	Reset(clientID string)
	StartRefillJob(ctx context.Context)
	Stop() error
}
