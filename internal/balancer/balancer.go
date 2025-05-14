package balancer

import "load-balancer/internal/backend"

type Balancer struct {
	strategy Strategy
	backends []*backend.Backend
}

func NewBalancer(strategy Strategy, backends []*backend.Backend) *Balancer {
	return &Balancer{
		strategy: strategy,
		backends: backends,
	}
}

func (b *Balancer) GetNext() string {
	return b.strategy.GetNext()
}

func (b *Balancer) GetBackends() []*backend.Backend {
	return b.backends
}
