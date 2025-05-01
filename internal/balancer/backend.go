package balancer

import "sync"

type Backend struct {
	Addr  string
	Alive bool

	mu sync.RWMutex
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

type BackendWithConnections struct {
	*Backend
	connections int
	mu          sync.Mutex
}

func (b *BackendWithConnections) Inc() {
	b.mu.Lock()
	b.connections++
	b.mu.Unlock()
}

func (b *BackendWithConnections) Dec() {
	b.mu.Lock()
	if b.connections > 0 {
		b.connections--
	}
	b.mu.Unlock()
}

func (b *BackendWithConnections) Connections() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.connections
}
