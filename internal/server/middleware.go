package server

import (
	"load-balancer/internal/client"
	"net/http"
	"sync"
	"time"
)

type LimiterManager struct {
	clientStore client.ClientStore
	limiters    map[string]*ClientLimiter
	mu          sync.Mutex
}

type ClientLimiter struct {
	mu         sync.Mutex
	capacity   int
	tokens     int
	ratePerSec int
	lastRefill time.Time
}

func NewLimiterManager(store client.ClientStore) *LimiterManager {
	return &LimiterManager{
		clientStore: store,
		limiters:    make(map[string]*ClientLimiter),
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

	clients, err := m.clientStore.List()
	if err != nil {
		return nil, err
	}
	for _, c := range clients {
		if c.APIKey == apiKey {
			limiter = newClientLimiter(c.Capacity, c.RatePerSec)
			m.limiters[apiKey] = limiter
			return limiter, nil
		}
	}
	return nil, http.ErrNoCookie
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
			http.Error(w, "Invalid API key", http.StatusForbidden)
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
