package balancer

import (
	"errors"
	"fmt"
	"sync"
)

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
		return "", errors.New("no available backends")
	}

	backend := b.backends[b.current]
	b.current = (b.current + 1) % len(b.backends)

	fmt.Println("Selected backend:", backend) // test log
	return backend, nil
}

func (b *RoundRobinBalancer) HealthCheck() error {
	return nil
}
