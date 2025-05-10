package balancer

import "sync"

type Balancer struct {
	backends []string
	index    int
	mutex    sync.Mutex
}

func NewBalancer(backends []string) *Balancer {
	return &Balancer{
		backends: backends,
	}
}

func (b *Balancer) GetNext() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.backends) == 0 {
		return ""
	}

	upstream := b.backends[b.index]
	b.index = (b.index + 1) % len(b.backends)
	return upstream
}
