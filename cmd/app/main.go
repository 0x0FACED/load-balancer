package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	appconfig "github.com/0x0FACED/load-balancer/config"
	"github.com/0x0FACED/load-balancer/internal/app"
	"github.com/0x0FACED/load-balancer/internal/balancer"
	"github.com/0x0FACED/load-balancer/internal/client"
	"github.com/0x0FACED/load-balancer/internal/limitter"
	"github.com/0x0FACED/load-balancer/internal/middleware"
	"github.com/0x0FACED/load-balancer/internal/server"
	"github.com/0x0FACED/zlog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := appconfig.Load()
	if err != nil {
		panic(err)
	}

	logger, err := zlog.NewZerologLogger(zlog.LoggerConfig{
		LogLevel: cfg.Logger.Level,
		LogsDir:  cfg.Logger.LogsDir,
	})
	if err != nil {
		log.Fatalln("Failed to initialize logger:", err)
	}

	logger.Info().Msg("Logger initialized")

	logger.Info().Any("config", cfg).Msg("Loaded configuration")

	appLogger := logger.ChildWithName("component", "app")
	middlewareLogger := logger.ChildWithName("component", "middleware")
	balancerLogger := logger.ChildWithName("component", "balancer")

	var bal balancer.Balancer
	switch cfg.Balancer.Type {
	case balancer.RoundRobin:
		bal = balancer.NewRoundRobinBalancer(cfg.Balancer.Backends, balancerLogger)
	case balancer.LeastConn:
		bal = balancer.NewLeastConnectionsBalancer(cfg.Balancer.Backends, balancerLogger)
	default:
		appLogger.Fatal().Msgf("Unknown balancer type: %s", cfg.Balancer.Type)
	}

	appLogger.Info().Msgf("Using %s balancer", cfg.Balancer.Type)

	db, err := sql.Open("postgres", cfg.Database.DSN)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to database")
	}

	clientRepo := client.NewPostgresRepo(db)

	limitter := limitter.NewTokenBucketLimitter(clientRepo, cfg.RateLimitter)

	loggerMiddleware := middleware.NewLoggerMiddleware(middlewareLogger)
	proxyMiddleware := middleware.NewProxyMiddleware(bal)
	limitterMiddleware := middleware.NewRateLimiterMiddleware(limitter)

	mux := http.NewServeMux()

	handler := loggerMiddleware.Logger(limitterMiddleware.Limitter(proxyMiddleware.Proxy(mux)))

	// основной сервер
	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Millisecond,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Millisecond,
	}

	muxReplica := http.NewServeMux()
	clientHandler := client.NewClientHandler(clientRepo)
	clientHandler.RegisterRoutes(muxReplica)

	muxReplica.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong from replica"))
	})

	// реплики серверы
	replicaServers := make([]*server.Server, len(cfg.Balancer.Backends))
	for i, backend := range cfg.Balancer.Backends {
		addr, err := extractHostPort(backend)
		if err != nil {
			logger.Error().Err(err).Msgf("invalid backend URL: %s", backend)
			continue
		}
		replicaServers[i] = server.New(
			&http.Server{
				Addr:         addr,
				Handler:      muxReplica,
				ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Millisecond,
				WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Millisecond,
				IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Millisecond,
			},
		)

	}

	app := app.New(srv, replicaServers, limitter, bal, appLogger, *cfg)

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
