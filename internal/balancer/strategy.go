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

	aliveBackends := make([]*backend.Backend, 0, len(r.backends))
	for _, b := range r.backends {
		if b.IsAlive() {
			aliveBackends = append(aliveBackends, b)
		}
	}

	if len(aliveBackends) == 0 {
		log.Warn().Msg("No available backends found")
		return ""
	}

	selected := aliveBackends[r.index]

	r.index = (r.index + 1) % len(aliveBackends)

	log.Debug().
		Str("selected_backend", selected.Addr).
		Msg("Selected backend for request")

	return selected.Addr
}
