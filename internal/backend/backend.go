package backend

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"

	"net/http"
)

type Backend struct {
	Addr  string
	Alive bool
}

func StartBackend(ctx context.Context, backends []*Backend) {
	for _, backend := range backends {
		go func(addr string) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Response from backend %s", addr)
			})
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			srv := &http.Server{
				Addr:    addr,
				Handler: mux,
			}

			go func() {
				log.Info().Str("addres", addr).Msg("Starting backend server")
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error().
						Err(err).
						Str("address", addr).
						Msg("Failed to start backend server")
				}
			}()

			<-ctx.Done()
			log.Info().Str("address", addr).Msg("Shutting down backend server")

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error().
					Err(err).
					Str("address", addr).
					Msg("Backend server shutdown failed")
			}
			log.Info().Str("address", addr).Msg("Backend server stopped")
		}(backend.Addr)
	}

	return
}
