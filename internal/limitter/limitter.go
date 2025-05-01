package limitter

import "context"

type RateLimitter interface {
	Allow(ctx context.Context, clientID string) bool
	Reset(clientID string)
	StartRefillJob(ctx context.Context)
	Stop() error
}
