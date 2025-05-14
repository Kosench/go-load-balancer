package main

import (
	"github.com/rs/zerolog/log"
	"load-balancer/internal/backend"
	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
	"load-balancer/internal/health"
	"load-balancer/internal/server"
	"time"
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
			Addr:  addr,
			Alive: true,
		})
	}

	backend.StartBackend(backends)

	strategy := balancer.NewRoundRobinStrategy(backends)
	lb := balancer.NewBalancer(strategy, backends)

	time.Sleep(2 * time.Second)
	health.StartHealthCheck(backends, 5*time.Second)

	srv := server.NewServer(cfg, lb)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}

}
