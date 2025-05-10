package balancer

import "sync"

type Strategy interface {
	GetNext() string
}

type RoundRobinStrategy struct {
	backends []string
	index    int
	mutex    sync.Mutex
}

func NewRoundRobinStrategy(backends []string) *RoundRobinStrategy {
	return &RoundRobinStrategy{
		backends: backends,
	}
}

func (r *RoundRobinStrategy) GetNext() string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.backends) == 0 {
		return ""
	}

	upstream := r.backends[r.index]
	r.index = (r.index + 1) % len(r.backends)
	return upstream
}
