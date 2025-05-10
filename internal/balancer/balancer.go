package balancer

type Balancer struct {
	strategy Strategy
}

func NewBalancer(strategy Strategy) *Balancer {
	return &Balancer{
		strategy: strategy,
	}
}

func (b *Balancer) GetNext() string {
	return b.strategy.GetNext()
}
