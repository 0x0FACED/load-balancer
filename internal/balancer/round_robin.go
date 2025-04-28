package balancer

import "sync"

type RoundRobinBalancer struct {
	Backends []string
	Current  int

	mu sync.Mutex
}

func NewRoundRobinBalancer(backends []string) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		Backends: backends,
		Current:  0,
	}
}

func (b *RoundRobinBalancer) Next() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.Backends) == 0 {
		return "", nil
	}

	backend := b.Backends[b.Current]
	b.Current = (b.Current + 1) % len(b.Backends)

	return backend, nil
}
func (b *RoundRobinBalancer) HealthCheck() error {
	return nil
}
