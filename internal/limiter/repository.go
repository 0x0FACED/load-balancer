package limiter

import (
	"context"

	"github.com/0x0FACED/load-balancer/internal/client"
)

type Repository interface {
	Get(ctx context.Context, id string) (*client.Client, error)
	Close() error
}
