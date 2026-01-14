package app

import (
	"context"
	"net/http"
	"time"

	"github.com/0x0FACED/load-balancer/config"
	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/load-balancer/internal/limiter"
	"github.com/0x0FACED/zlog"
	"go.uber.org/multierr"
)

type App struct {
	srv      *http.Server
	limiter  limiter.RateLimitter
	balancer balancer.Balancer

	cfg config.AppConfig
	log *zlog.ZerologLogger
}

func New(
	srv *http.Server,
	limitter limiter.RateLimitter,
	balancer balancer.Balancer,
	log *zlog.ZerologLogger,
	cfg config.AppConfig,
) *App {
	return &App{
		srv:      srv,
		limiter:  limitter,
		balancer: balancer,
		log:      log,
		cfg:      cfg,
	}
}

func (a *App) Start(ctx context.Context) error {
	errChan := make(chan error, 3)

	go func() {
		a.log.Info().Str("address", a.srv.Addr).Msg("Starting application server")
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	a.log.Info().Msg("Starting rate limiter refill job")
	a.limiter.StartRefillJob(ctx)

	a.log.Info().Msg("Starting health check job")
	a.balancer.StartHealthCheckJob(ctx)

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.log.Info().Msg("Shutting down servers...")

	var retErr error

	if err := a.srv.Shutdown(ctx); err != nil {
		a.log.Error().Err(err).Msg("Failed to shutdown application server")
		retErr = multierr.Append(retErr, err)
	} else {
		a.log.Info().Msg("Application server stopped")
	}

	if err := a.limiter.Stop(); err != nil {
		a.log.Error().Err(err).Msg("Failed to stop rate limiter")
		retErr = multierr.Append(retErr, err)
	} else {
		a.log.Info().Msg("Rate limiter stopped")
	}

	return retErr
}
