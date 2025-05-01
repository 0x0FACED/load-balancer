package balancer

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/0x0FACED/zlog"
)

type LeastConnectionsBalancer struct {
	backends []*BackendWithConnections
	log      *zlog.ZerologLogger
	mu       sync.Mutex
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

func NewLeastConnectionsBalancer(backends []string, log *zlog.ZerologLogger) *LeastConnectionsBalancer {
	backendsWithConnections := make([]*BackendWithConnections, len(backends))
	for i, addr := range backends {
		backendsWithConnections[i] = &BackendWithConnections{
			Backend: &Backend{
				Addr:  addr,
				Alive: true,
			},
			connections: 0,
		}
	}

	return &LeastConnectionsBalancer{
		backends: backendsWithConnections,
		log:      log,
	}
}

func (b *LeastConnectionsBalancer) Next() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var selected *BackendWithConnections
	for _, backend := range b.backends {
		if !backend.IsAlive() {
			continue
		}
		if selected == nil || backend.Connections() < selected.Connections() {
			selected = backend
		}
	}

	if selected == nil {
		return "", ErrNoBackends
	}

	selected.Inc()

	b.log.Debug().Str("addr", selected.Addr).Int("connections", selected.Connections()).Msg("[LeastConn] is selected")

	return selected.Addr, nil
}

func (b *LeastConnectionsBalancer) Release(addr string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, backend := range b.backends {
		if backend.Addr == addr {
			backend.Dec()
			b.log.Debug().Str("addr", addr).Int("connections", backend.Connections()).Msg("[LeastConn] released")
			return
		}
	}

	b.log.Warn().Str("addr", addr).Msg("[LeastConn] release failed: backend not found")
}

func (b *LeastConnectionsBalancer) StartHealthCheckJob(ctx context.Context) {
	for _, backend := range b.backends {
		go func(bk *BackendWithConnections) {
			client := &http.Client{Timeout: healthCheckTimeout}

			ticker := time.NewTicker(healthCheckInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					b.log.Info().Str("addr", bk.Addr).Msg("[HealthCheck] stopped")
					return
				case <-ticker.C:
					url := bk.Addr + "/ping"
					resp, err := client.Get(url)
					if err != nil {
						b.log.Error().Str("addr", bk.Addr).Err(err).Msg("[HealthCheck] is DOWN")
						bk.SetAlive(false)
						continue
					}

					_ = resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						b.log.Debug().Str("addr", bk.Addr).Int("connections", bk.connections).Int("code", resp.StatusCode).Msg("[HealthCheck] is UP")
						bk.SetAlive(true)
					} else {
						b.log.Warn().Str("addr", bk.Addr).Int("code", resp.StatusCode).Msg("[HealthCheck] not ready")
						bk.SetAlive(false)
					}
				}
			}
		}(backend)
	}
}
