package main

import (
	"github.com/rs/zerolog/log"
	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
	"load-balancer/internal/server"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	strategy := balancer.NewRoundRobinStrategy(cfg.Backends)

	lb := balancer.NewBalancer(strategy)

	srv := server.NewServer(cfg, lb)
	if err := srv.Start(); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}

}
