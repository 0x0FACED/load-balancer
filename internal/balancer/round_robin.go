package balancer

import "sync"

type RoundRobinBalancer struct {
	backends []string
	current  int

	mu sync.Mutex
}

func NewRoundRobinBalancer(backends []string) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		backends: backends,
		current:  0,
	}
}

func (b *RoundRobinBalancer) Next() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.backends) == 0 {
		return "", nil
	}

	backend := b.backends[b.current]
	b.current = (b.current + 1) % len(b.backends)

	return backend, nil
}
func (b *RoundRobinBalancer) HealthCheck() error {
	return nil
}
