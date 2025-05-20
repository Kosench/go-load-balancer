package backend

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"sync"
	"sync/atomic"
	"time"

	"net/http"
)

type Backend struct {
	Addr  string
	alive int32
}

func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.alive) == 1
}

func (b *Backend) SetAlive(state bool) {
	var value int32
	if state {
		value = 1
	}
	atomic.StoreInt32(&b.alive, value)
}

func StartBackend(ctx context.Context, backends []*Backend) {
	var wg sync.WaitGroup
	for _, b := range backends {
		wg.Add(1)

		go func(b *Backend) {
			defer wg.Done()

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Response from backend %s", b.Addr)
			})
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			srv := &http.Server{
				Addr:    b.Addr,
				Handler: mux,
			}

			go func() {
				log.Info().Str("addres", b.Addr).Msg("Starting backend server")
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error().
						Err(err).
						Str("address", b.Addr).
						Msg("Failed to start backend server")
				}
			}()

			<-ctx.Done()
			log.Info().Str("address", b.Addr).Msg("Shutting down backend server")

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error().
					Err(err).
					Str("address", b.Addr).
					Msg("Backend server shutdown failed")
			}
			log.Info().Str("address", b.Addr).Msg("Backend server stopped")
		}(b)
	}

	wg.Wait()
}
