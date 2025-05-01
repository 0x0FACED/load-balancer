package balancer

import "context"

type BalancerType string

const (
	RoundRobin BalancerType = "round_robin"
	LeastConn  BalancerType = "least_conn"
)

type Balancer interface {
	Next() (string, error)
	Release(addr string)
	StartHealthCheckJob(ctx context.Context)
}
