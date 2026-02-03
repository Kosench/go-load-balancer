// Package backend provides backend server management functionality.
package backend

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	"net/http"
)

// Backend represents a single backend server with health status tracking.
// The Alive field uses atomic operations for thread-safe access.
type Backend struct {
	Addr  string // Address of the backend server (host:port)
	Alive int32  // Health status: 1 for alive, 0 for dead (accessed atomically)
}

// IsAlive returns true if the backend is currently healthy and available.
// Thread-safe using atomic load operation.
func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.Alive) == 1
}

// SetAlive updates the health status of the backend.
// Thread-safe using atomic store operation.
func (b *Backend) SetAlive(state bool) {
	var value int32
	if state {
		value = 1
	}
	atomic.StoreInt32(&b.Alive, value)
}

// StartBackend starts HTTP servers for each backend in separate goroutines.
// Each backend server provides a simple echo endpoint and a health check endpoint.
// The function signals via the ready channel once all backends are listening and ready.
//
// Parameters:
//   - ctx: Context for graceful shutdown
//   - backends: Slice of backend servers to start
//   - wg: WaitGroup to coordinate shutdown
//   - ready: Channel to signal when all backends are ready (will be closed)
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

			<-ctx.Done()
			log.Info().Str("address", b.Addr).Msg("Shutting down backend server")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error().Err(err).Str("address", b.Addr).Msg("Backend server shutdown failed")
			}
			log.Info().Str("address", b.Addr).Msg("Backend server stopped")
		}(b)
	}
}
