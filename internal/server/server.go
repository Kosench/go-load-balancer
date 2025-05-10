package server

import (
	"load-balancer/internal/balancer"
	"load-balancer/internal/config"
	"load-balancer/pkg/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server struct {
	Config   *config.Config
	Balancer *balancer.Balancer
	Logger   *utils.Logger
}

func NewServer(cfg *config.Config) *Server {
	logger := utils.NewLogger(cfg.LogLevel)
	return &Server{
		Config:   cfg,
		Balancer: balancer.NewBalancer(cfg.Backends),
		Logger:   logger,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleRequest)
	s.Logger.Info("Starting server on %s", s.Config.ListenAddress)
	return http.ListenAndServe(s.Config.ListenAddress, nil)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	s.Logger.Debug("Incoming request: %s %s", r.Method, r.URL.Path)

	upstream := s.Balancer.GetNext()
	if upstream == "" {
		s.Logger.Error("No available backends")
		http.Error(w, "No available backends", http.StatusServiceUnavailable)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: upstream})
	proxy.ServeHTTP(w, r)
}
