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
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping health checks")
				return
			case <-ticker.C:
				for _, b := range backends {
					go checkBackend(b)
				}
			}
		}
	}()
}

func checkBackend(b *backend.Backend) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("http://" + b.Addr + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		b.SetAlive(false)
	} else {
		b.SetAlive(true)
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Log the health status
	log.Info().
		Str("backend", b.Addr).
		Bool("alive", b.IsAlive()).
		Msg("Backend health status updated")
}
