package backend

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"net/http"
)

type Backend struct {
	Addr  string
	Alive int32
}

func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.Alive) == 1
}

func (b *Backend) SetAlive(state bool) {
	var value int32
	if state {
		value = 1
	}
	atomic.StoreInt32(&b.Alive, value)
}

func StartBackend(ctx context.Context, backends []*Backend, wg *sync.WaitGroup, ready chan<- struct{}) {
	readyCount := 0
	readyMu := sync.Mutex{}
	totalBackends := len(backends)

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

			var listener net.Listener
			var err error
			addr := b.Addr
			if addr == "" || addr == ":0" {
				listener, err = net.Listen("tcp", ":0")
				if err != nil {
					log.Error().Err(err).Msg("Failed to listen on a random port")
					b.SetAlive(false)
					return
				}
				b.Addr = listener.Addr().String()
			} else {
				listener, err = net.Listen("tcp", addr)
				if err != nil {
					log.Error().Err(err).Str("address", addr).Msg("Failed to listen")
					b.SetAlive(false)
					return
				}
			}

			srv := &http.Server{
				Handler: mux,
			}

			serverDone := make(chan struct{})
			defer close(serverDone)

			// Set alive only after successful listen
			b.SetAlive(true)
			log.Info().Str("address", b.Addr).Msg("Starting backend server")

			// Notify that this backend is ready
			readyMu.Lock()
			readyCount++
			if readyCount == totalBackends && ready != nil {
				close(ready)
			}
			readyMu.Unlock()

			go func() {
				if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
					log.Error().Err(err).Str("address", b.Addr).Msg("Failed to start backend server")
				}
			}()

			select {
			case <-ctx.Done():
				log.Info().Str("address", b.Addr).Msg("Shutting down backend server")
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.Shutdown(shutdownCtx); err != nil {
					log.Error().Err(err).Str("address", b.Addr).Msg("Backend server shutdown failed")
				}
				log.Info().Str("address", b.Addr).Msg("Backend server stopped")
			}
		}(b)
	}
}
