package balancer

type Balancer interface {
	Next() (string, error)
	HealthCheck() error
}
