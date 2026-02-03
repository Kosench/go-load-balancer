// Package balancer provides load balancing functionality with pluggable strategies.
package balancer

import "load-balancer/internal/backend"

// Balancer distributes incoming requests across multiple backend servers
// using a configurable balancing strategy.
type Balancer struct {
	strategy Strategy
	backends []*backend.Backend
}

// NewBalancer creates a new load balancer with the given strategy and backends.
func NewBalancer(strategy Strategy, backends []*backend.Backend) *Balancer {
	return &Balancer{
		strategy: strategy,
		backends: backends,
	}
}

// GetNext returns the address of the next backend to use for a request,
// according to the configured balancing strategy.
// Returns an empty string if no backends are available.
func (b *Balancer) GetNext() string {
	return b.strategy.GetNext()
}

// GetBackends returns the list of all backends managed by this balancer.
func (b *Balancer) GetBackends() []*backend.Backend {
	return b.backends
}
