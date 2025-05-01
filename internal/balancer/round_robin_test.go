package balancer_test

import (
	"testing"

	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/zlog"
	"github.com/stretchr/testify/assert"
)

func TestRoundRobinBalancer_Next(t *testing.T) {
	log := zlog.NewTestLogger()
	backends := []string{"a", "b", "c"}
	cfg := balancer.Config{
		Backends: backends,
	}
	b := balancer.NewRoundRobinBalancer(log, cfg)

	b.SetAlive("a", true)
	b.SetAlive("b", true)
	b.SetAlive("c", true)

	addr1, _ := b.Next()
	addr2, _ := b.Next()
	addr3, _ := b.Next()
	addr4, _ := b.Next()

	assert.Equal(t, "a", addr1)
	assert.Equal(t, "b", addr2)
	assert.Equal(t, "c", addr3)
	assert.Equal(t, "a", addr4)
}
