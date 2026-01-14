package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	appconfig "github.com/0x0FACED/load-balancer/config"
	"github.com/0x0FACED/load-balancer/internal/app"
	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/load-balancer/internal/client"
	"github.com/0x0FACED/load-balancer/internal/limiter"
	"github.com/0x0FACED/load-balancer/internal/middleware"
	"github.com/0x0FACED/zlog"
)

func main() {
	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// app config load
	cfg, err := appconfig.Load()
	if err != nil {
		panic(err)
	}

	// logger init
	logger, err := zlog.NewZerologLogger(zlog.LoggerConfig{
		LogLevel: cfg.Logger.Level,
		LogsDir:  cfg.Logger.LogsDir,
	})
	if err != nil {
		log.Fatalln("Failed to initialize logger:", err)
	}

	logger.Info().Msg("Logger initialized")

	logger.Info().Any("config", cfg).Msg("Loaded configuration")

	// creating logger for different components
	appLogger := logger.ChildWithName("component", "app")
	middlewareLogger := logger.ChildWithName("component", "middleware")
	balancerLogger := logger.ChildWithName("component", "balancer")

	// init balancer
	var bal balancer.Balancer
	switch cfg.Balancer.Type {
	case balancer.RoundRobin:
		bal = balancer.NewRoundRobinBalancer(balancerLogger, cfg.Balancer)
	case balancer.LeastConn:
		bal = balancer.NewLeastConnectionsBalancer(balancerLogger, cfg.Balancer)
	default:
		appLogger.Fatal().Msgf("Unknown balancer type: %s", cfg.Balancer.Type)
	}

	appLogger.Info().Msgf("Using %s balancer", cfg.Balancer.Type)

	// init database
	db, err := sql.Open("postgres", cfg.Database.DSN)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// TODO: remove
	clientRepo := client.NewPostgresRepo(db)

	limiter := limiter.NewTokenBucketLimiter(clientRepo, cfg.RateLimiter)

	loggerMiddleware := middleware.NewLoggerMiddleware(middlewareLogger)
	proxyMiddleware := middleware.NewProxyMiddleware(bal)
	limitterMiddleware := middleware.NewRateLimiterMiddleware(limiter)

	mux := http.NewServeMux()

	handler := loggerMiddleware.Logger(limitterMiddleware.Limiter(proxyMiddleware.Proxy(mux)))

	// основной сервер
	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Millisecond,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Millisecond,
	}

	app := app.New(srv, limiter, bal, appLogger, *cfg)

	go func() {
		if err := app.Start(ctx); err != nil {
			return
		}
	}()

	<-ctx.Done()

	if err := app.Shutdown(); err != nil {
		return
	}

}

func extractHostPort(backendURL string) (string, error) {
	u, err := url.Parse(backendURL)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}
