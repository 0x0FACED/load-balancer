package balancer

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/0x0FACED/zlog"
)

var (
	ErrNoBackends      = errors.New("no backends available")
	ErrBackendNotAlive = errors.New("backend is not alive")
)

type RoundRobinBalancer struct {
	backends []*Backend
	current  int

	log *zlog.ZerologLogger

	mu sync.Mutex
}

func NewRoundRobinBalancer(backends []string, log *zlog.ZerologLogger) *RoundRobinBalancer {
	backendsList := make([]*Backend, len(backends))
	for i, addr := range backends {
		backendsList[i] = &Backend{
			Addr:  addr,
			Alive: true,
		}
	}

	return &RoundRobinBalancer{
		backends: backendsList,
		current:  0,
		log:      log,
	}
}

func (b *RoundRobinBalancer) Next() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.backends) == 0 {
		b.log.Error().Msg("No backends available")
		return "", ErrNoBackends
	}

	var counter int
	for range b.backends {
		if counter > len(b.backends) {
			return "", ErrNoBackends
		}
		if b.backends[b.current].IsAlive() {
			break
		}
		b.current = (b.current + 1) % len(b.backends)
		counter++
	}
	backend := b.backends[b.current]

	b.log.Debug().Str("addr", backend.Addr).Msg("[Next] is selected")
	// double check
	if !backend.IsAlive() {
		return "", ErrBackendNotAlive
	}
	b.current = (b.current + 1) % len(b.backends)

	return backend.Addr, nil
}

// заглушка
func (b *RoundRobinBalancer) Release(addr string) {
	return
}

const healthCheckInterval = 3 * time.Second
const healthCheckTimeout = 2 * time.Second

func (b *RoundRobinBalancer) StartHealthCheckJob(ctx context.Context) {
	for _, backend := range b.backends {
		go func(bk *Backend) {
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
						b.log.Debug().Str("addr", bk.Addr).Int("code", resp.StatusCode).Msg("[HealthCheck] is UP")
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
