package backend

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"load-balancer/internal/config"
	"net/http"
)

func StartBackend(cfg *config.Config) {
	for _, backend := range cfg.Backends {
		go func(addr string) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Response from backend %s", addr)
			})

			log.Printf("Starting backend server on %s", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Fatal().
					Err(err).
					Str("address", addr).
					Msg("Failed to start backend server on")
			}

		}(backend)
	}

	return
}
