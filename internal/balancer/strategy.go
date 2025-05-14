package balancer

import (
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

	n := len(r.backends)
	for i := 0; i < n; i++ {
		r.index = (r.index + 1) % n
		if r.backends[r.index].Alive {
			return r.backends[r.index].Addr
		}
	}
	return ""
}
