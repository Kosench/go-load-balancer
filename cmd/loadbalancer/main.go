package main

import (
	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
	"load-balancer/internal/server"
	"log"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	strategy := balancer.NewRoundRobinStrategy(cfg.Backends)

	lb := balancer.NewBalancer(strategy)

	srv := server.NewServer(cfg, lb)
	if err := srv.Start(); err != nil {
		log.Fatal("Server failed: %v", err)
	}

}
