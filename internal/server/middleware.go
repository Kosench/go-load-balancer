package server

import (
	"errors"
	"load-balancer/internal/client"
	"load-balancer/internal/config"
	"net/http"
	"sync"
	"time"
)

type LimiterManager struct {
	clientStore    client.ClientStore
	limiters       map[string]*ClientLimiter
	mu             sync.Mutex
	defaultConfig  *config.Config
}

type ClientLimiter struct {
	mu         sync.Mutex
	capacity   int
	tokens     int
	ratePerSec int
	lastRefill time.Time
}

func NewLimiterManager(store client.ClientStore, cfg *config.Config) *LimiterManager {
	return &LimiterManager{
		clientStore:   store,
		limiters:      make(map[string]*ClientLimiter),
		defaultConfig: cfg,
	}
}

func newClientLimiter(capacity int, ratePerSec int) *ClientLimiter {
	return &ClientLimiter{
		capacity:   capacity,
		tokens:     capacity,
		ratePerSec: ratePerSec,
		lastRefill: time.Now(),
	}
}

func (m *LimiterManager) GetLimiter(apiKey string) (*ClientLimiter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	limiter, ok := m.limiters[apiKey]
	if ok {
		return limiter, nil
	}

	// Use GetByAPIKey instead of iterating through List()
	c, err := m.clientStore.GetByAPIKey(apiKey)
	if err != nil {
		if errors.Is(err, client.ErrClientNotFound) {
			return nil, client.ErrClientNotFound
		}
		return nil, err
	}

	limiter = newClientLimiter(c.Capacity, c.RatePerSec)
	m.limiters[apiKey] = limiter
	return limiter, nil
}

func (m *LimiterManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}
		limiter, err := m.GetLimiter(apiKey)
		if err != nil {
			if errors.Is(err, client.ErrClientNotFound) {
				http.Error(w, "Invalid API key", http.StatusForbidden)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *ClientLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	newTokens := int(elapsed * float64(l.ratePerSec))
	if newTokens > 0 {
		l.tokens = min(l.capacity, l.tokens+newTokens)
		l.lastRefill = now
	}

	if l.tokens > 0 {
		l.tokens--
		return true
	}
	return false
}
