package health

import (
	"github.com/rs/zerolog/log"
	"load-balancer/internal/backend"
	"net/http"
	"time"
)

func StartHealthCheck(backends []*backend.Backend, interval time.Duration) {
	go func() {
		for {
			for _, b := range backends {
				go checkBackend(b)
			}
			time.Sleep(interval)
		}
	}()
}

func checkBackend(b *backend.Backend) {
	resp, err := http.Get("http://" + b.Addr + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		b.SetAlive(false)
	} else {
		b.SetAlive(true)
	}
	if resp != nil {
		resp.Body.Close()
	}

	log.Info().
		Str("backend", b.Addr).
		Bool("alive", b.IsAlive()).
		Msg("Backend health status updated")
}
