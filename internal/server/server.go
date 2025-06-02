package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"load-balancer/internal/balancer"
	"load-balancer/internal/client"
	"load-balancer/internal/config"
	"load-balancer/internal/ratelimit"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Config        *config.Config
	Balancer      *balancer.Balancer
	RateLimiter   ratelimit.RateLimiter
	srv           *http.Server
	proxies       map[string]*httputil.ReverseProxy
	proxiesMu     sync.RWMutex
	clientHandler *client.Handler
}

func NewServer(cfg *config.Config, lb *balancer.Balancer) *Server {
	proxies := make(map[string]*httputil.ReverseProxy)
	for _, backend := range lb.GetBackends() {
		url, err := url.Parse("http://" + backend.Addr)
		if err != nil {
			log.Error().Err(err).Str("backend", backend.Addr).Msg("Failed to parse backend URL")
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		// Настройка таймаутов для прокси
		proxy.Transport = &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       30 * time.Second,
		}
		proxies[backend.Addr] = proxy
	}

	clientStore := client.NewInMemoryClientStore()
	clientHandler := client.NewHandler(clientStore)
	clientMux := http.NewServeMux()
	clientHandler.RegisterRoutes(clientMux)

	limiterManager := NewLimiterManager(clientStore)

	server := &Server{
		Config:   cfg,
		Balancer: lb,
		RateLimiter: ratelimit.NewTokenBucketRateLimiter(ratelimit.RateLimiterConfig{
			Capacity:   cfg.RateLimitCapacity,
			RefillRate: cfg.RateLimitRefillRate,
		}),
		proxies:       proxies,
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

func (s *Server) Start() error {
	log.Info().Msgf("Starting server on %s", s.Config.ListenAddress)
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down server")
	return s.srv.Shutdown(ctx)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Error().Err(err).Str("remote_addr", r.RemoteAddr).Msg("Failed to parse client IP")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	clientID := host // Используем только IP как идентификатор клиента

	if !s.RateLimiter.Allow(clientID) {
		log.Error().
			Str("client", clientID).
			Msg("Request rate limited")
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

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

	proxy, exists := s.proxies[upstream]
	if !exists {
		log.Error().Str("backend", upstream).Msg("No proxy found for backend")
		http.Error(w, "No proxy found for backend", http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("backend", upstream).
		Str("path", r.URL.Path).
		Msg("Proxying request to backend")

	proxy.ServeHTTP(w, r)
}
