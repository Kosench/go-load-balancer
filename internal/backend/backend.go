package backend

import (
	"fmt"
	"github.com/rs/zerolog/log"

	"net/http"
)

type Backend struct {
	Addr  string
	Alive bool
}

func StartBackend(backends []*Backend) {
	for _, backend := range backends {
		go func(addr string) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Response from backend %s", addr)
			})
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			log.Info().Str("address", addr).Msg("Starting backend server")
			if err := http.ListenAndServe(addr, mux); err != nil {
				log.Fatal().
					Err(err).
					Str("address", addr).
					Msg("Failed to start backend server on")
			}

		}(backend.Addr)
	}

	return
}
