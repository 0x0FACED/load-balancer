package balancer

import "context"

type Balancer interface {
	Next() (string, error)
	StartHealthCheckJob(ctx context.Context)
}
