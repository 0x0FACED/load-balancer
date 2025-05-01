package balancer_test

import (
	"testing"

	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/zlog"
	"github.com/stretchr/testify/assert"
)

func TestLeastConnectionsBalancer_Next(t *testing.T) {
	log := zlog.NewTestLogger()
	backends := []string{"a", "b"}
	cfg := balancer.Config{
		Backends: backends,
	}
	b := balancer.NewLeastConnectionsBalancer(log, cfg)

	b.SetAlive("a", true)
	b.SetAlive("b", true)
	// a conns = 1
	// b conns = 0
	addr1, err := b.Next()
	assert.NoError(t, err)
	assert.Contains(t, backends, addr1)
	assert.Equal(t, "a", addr1)

	// a conns = 1
	// b conns = 1
	addr2, err := b.Next()
	assert.NoError(t, err)
	assert.Contains(t, backends, addr2)
	assert.Equal(t, "b", addr2)

	b.Release(addr1)

	// a conns = 0
	// a conns = 1
	addr3, err := b.Next()
	assert.NoError(t, err)
	assert.Contains(t, backends, addr3)
	assert.Equal(t, "a", addr3)
}
