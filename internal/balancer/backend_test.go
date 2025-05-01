package balancer_test

import (
	"sync"
	"testing"

	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/stretchr/testify/assert"
)

func TestBackend_SetAndIsAlive(t *testing.T) {
	b := &balancer.Backend{Addr: "http://localhost:8080"}

	assert.False(t, b.IsAlive(), "backend should not be alive by default")

	b.SetAlive(true)
	assert.True(t, b.IsAlive(), "backend should be alive after SetAlive(true)")

	b.SetAlive(false)
	assert.False(t, b.IsAlive(), "backend should be not alive after SetAlive(false)")
}

func TestBackend_ConcurrentAliveAccess(t *testing.T) {
	b := &balancer.Backend{Addr: "http://localhost:8080"}
	var wg sync.WaitGroup
	setCount := 100

	for i := 0; i < setCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b.SetAlive(i%2 == 0)
		}(i)
	}

	wg.Wait()
	_ = b.IsAlive() // ensure no race
}

func TestBackendWithConnections_IncDecConnections(t *testing.T) {
	b := &balancer.BackendWithConnections{
		Backend: &balancer.Backend{Addr: "http://localhost:8080"},
	}

	assert.Equal(t, 0, b.Connections(), "initial connections should be 0")

	b.Inc()
	assert.Equal(t, 1, b.Connections(), "connections should be 1 after Inc")

	b.Inc()
	assert.Equal(t, 2, b.Connections(), "connections should be 2 after second Inc")

	b.Dec()
	assert.Equal(t, 1, b.Connections(), "connections should be 1 after Dec")

	b.Dec()
	assert.Equal(t, 0, b.Connections(), "connections should be 0 after second Dec")

	b.Dec()
	assert.Equal(t, 0, b.Connections(), "connections should not go below 0")
}

func TestBackendWithConnections_ConcurrentIncDec(t *testing.T) {
	b := &balancer.BackendWithConnections{
		Backend: &balancer.Backend{Addr: "http://localhost:8080"},
	}

	var wg sync.WaitGroup
	count := 1000

	// concurrent increments
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Inc()
		}()
	}

	wg.Wait()
	assert.Equal(t, count, b.Connections(), "all concurrent Increments should be counted")

	// concurrent decrements
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Dec()
		}()
	}

	wg.Wait()
	assert.Equal(t, 0, b.Connections(), "all concurrent Decrements should reach 0")
}
