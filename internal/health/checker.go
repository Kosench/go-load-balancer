package health

import (
	"context"
	"github.com/rs/zerolog/log"
	"load-balancer/internal/backend"
	"net/http"
	"time"
)

func StartHealthCheck(ctx context.Context, backends []*backend.Backend, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Health checks stopped")
				return
			case <-ticker.C:
				for _, b := range backends {
					go checkBackend(b)
				}
			}
		}
	}()
}

var healthCheckClient = &http.Client{
	Timeout: 5 * time.Second,
}

func checkBackend(b *backend.Backend) {
	resp, err := healthCheckClient.Get("http://" + b.Addr + "/health")
	if err != nil {
		b.SetAlive(false)
		log.Info().
			Str("backend", b.Addr).
			Bool("alive", false).
			Err(err).
			Msg("Backend health check failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		b.SetAlive(true)
	} else {
		b.SetAlive(false)
	}

	// Log the health status
	log.Info().
		Str("backend", b.Addr).
		Bool("alive", b.IsAlive()).
		Int("status_code", resp.StatusCode).
		Msg("Backend health status updated")
}
