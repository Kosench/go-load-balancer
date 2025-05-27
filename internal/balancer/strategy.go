package balancer

import (
	"github.com/rs/zerolog/log"
	"load-balancer/internal/backend"
	"sync"
)

type Strategy interface {
	GetNext() string
}

type RoundRobinStrategy struct {
	backends []*backend.Backend
	index    int
	mutex    sync.Mutex
}

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
