// Package server provides HTTP server implementation with reverse proxy and rate limiting.
package server

import (
	"context"
	"load-balancer/internal/balancer"
	"load-balancer/internal/client"
	"load-balancer/internal/config"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Server represents the HTTP load balancer server.
// It handles incoming requests, applies rate limiting, and proxies to backends.
type Server struct {
	Config        *config.Config
	Balancer      *balancer.Balancer
	srv           *http.Server
	proxies       map[string]*httputil.ReverseProxy // Cached reverse proxies per backend
	proxiesMu     sync.RWMutex
	clientHandler *client.Handler
}

// NewServer creates a new load balancer server with the given configuration and balancer.
func NewServer(cfg *config.Config, lb *balancer.Balancer) *Server {
	clientStore := client.NewInMemoryClientStore()
	clientHandler := client.NewHandler(clientStore)
	clientMux := http.NewServeMux()
	clientHandler.RegisterRoutes(clientMux)

	limiterManager := NewLimiterManager(clientStore, cfg)

	server := &Server{
		Config:        cfg,
		Balancer:      lb,
		proxies:       make(map[string]*httputil.ReverseProxy),
		clientHandler: clientHandler,
	}

	proxyHandler := limiterManager.Middleware(http.HandlerFunc(server.handleRequest))

	server.srv = &http.Server{
		Addr: cfg.ListenAddress,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/clients") {
				clientMux.ServeHTTP(w, r)
				return
			}
			proxyHandler.ServeHTTP(w, r)
		}),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server
}

// Start starts the HTTP server and begins accepting requests.
func (s *Server) Start() error {
	log.Info().Msgf("Starting server on %s", s.Config.ListenAddress)
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server within the given context timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down server")
	return s.srv.Shutdown(ctx)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("method", r.Method).
		Stringer("url", r.URL).
		Msg("Incoming request")

	upstream := s.Balancer.GetNext()
	if upstream == "" {
		log.Error().Msg("No available backends")
		http.Error(w, "No available backends", http.StatusServiceUnavailable)
		return
	}

	proxy := s.getOrCreateProxy(upstream)
	if proxy == nil {
		log.Error().Str("backend", upstream).Msg("Failed to create proxy for backend")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("backend", upstream).
		Str("path", r.URL.Path).
		Msg("Proxying request to backend")

	proxy.ServeHTTP(w, r)
}

func (s *Server) getOrCreateProxy(backend string) *httputil.ReverseProxy {
	s.proxiesMu.RLock()
	proxy, exists := s.proxies[backend]
	s.proxiesMu.RUnlock()

	if exists {
		return proxy
	}

	s.proxiesMu.Lock()
	defer s.proxiesMu.Unlock()

	// Double-check in case another goroutine created it
	if proxy, exists := s.proxies[backend]; exists {
		return proxy
	}

	targetURL, err := url.Parse("http://" + backend)
	if err != nil {
		log.Error().Err(err).Str("backend", backend).Msg("Failed to parse backend URL")
		return nil
	}

	proxy = httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       30 * time.Second,
	}

	// Add error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().
			Err(err).
			Str("backend", backend).
			Str("path", r.URL.Path).
			Msg("Proxy error")
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	s.proxies[backend] = proxy
	log.Info().Str("backend", backend).Msg("Created new proxy for backend")

	return proxy
}
