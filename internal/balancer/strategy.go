package balancer

import (
	"load-balancer/internal/backend"
	"sync"

	"github.com/rs/zerolog/log"
)

// Strategy defines the interface for load balancing strategies.
// Implementations must be thread-safe.
type Strategy interface {
	// GetNext returns the address of the next backend to use.
	// Returns empty string if no backends are available.
	GetNext() string
}

// RoundRobinStrategy implements a round-robin load balancing strategy.
// It distributes requests evenly across all healthy backends in a circular manner.
type RoundRobinStrategy struct {
	backends []*backend.Backend
	index    int
	mutex    sync.Mutex
}

// NewRoundRobinStrategy creates a new round-robin strategy for the given backends.
func NewRoundRobinStrategy(backends []*backend.Backend) *RoundRobinStrategy {
	return &RoundRobinStrategy{
		backends: backends,
	}
}

func (r *RoundRobinStrategy) GetNext() string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	numBackends := len(r.backends)
	if numBackends == 0 {
		log.Warn().Msg("No backends configured")
		return ""
	}

	startIndex := r.index
	for i := 0; i < numBackends; i++ {
		idx := (startIndex + i) % numBackends
		if r.backends[idx].IsAlive() {
			r.index = (idx + 1) % numBackends
			log.Debug().
				Str("selected_backend", r.backends[idx].Addr).
				Msg("Selected backend for request")
			return r.backends[idx].Addr
		}
	}
	log.Warn().Msg("No available backends found")
	return ""
}
