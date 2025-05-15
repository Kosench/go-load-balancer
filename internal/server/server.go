package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server struct {
	Config   *config.Config
	Balancer *balancer.Balancer
	srv      *http.Server
}

func NewServer(cfg *config.Config, lb *balancer.Balancer) *Server {
	return &Server{
		Config:   cfg,
		Balancer: lb,
		srv: &http.Server{
			Addr: cfg.ListenAddress,
		},
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleRequest)
	log.Info().Msgf("Starting server on %s", s.Config.ListenAddress)
	return s.srv.ListenAndServe()
}

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

	log.Info().
		Str("backend", upstream).
		Str("path", r.URL.Path).
		Msg("Proxying request to backend")

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: upstream})
	proxy.ServeHTTP(w, r)
}
