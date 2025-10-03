package main

import (
	"context"
	"load-balancer/internal/backend"
	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
	"load-balancer/internal/health"
	"load-balancer/internal/server"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}
	log.Print(cfg)

	var backends []*backend.Backend
	for _, addr := range cfg.Backends {
		backends = append(backends, &backend.Backend{
			Addr: addr,
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	backendsReady := make(chan struct{})
	backend.StartBackend(ctx, backends, &wg, backendsReady)

	// Wait for all backends to be ready
	<-backendsReady
	log.Info().Msg("All backends are ready")

	strategy := balancer.NewRoundRobinStrategy(backends)
	lb := balancer.NewBalancer(strategy, backends)

	health.StartHealthCheck(ctx, backends, 15*time.Second)

	srv := server.NewServer(cfg, lb)
	go func() {
		err := srv.Start()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Info().Msg("Received shutdown signal, initiating graceful shutdown")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}

	wg.Wait()

	log.Info().Msg("Server and backends stopped")
}
