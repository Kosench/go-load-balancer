package balancer

import (
	"load-balancer/internal/backend"
	"testing"
)

func TestRoundRobinStrategy_GetNext(t *testing.T) {
	backends := []*backend.Backend{
		{Addr: "localhost:9001", Alive: 1},
		{Addr: "localhost:9002", Alive: 1},
		{Addr: "localhost:9003", Alive: 0}, // Dead backend
	}
	strategy := NewRoundRobinStrategy(backends)

	expectedOrder := []string{"localhost:9001", "localhost:9002", "localhost:9001"}
	for i, expected := range expectedOrder {
		actual := strategy.GetNext()
		if actual != expected {
			t.Errorf("Test %d: expected %s, got %s", i, expected, actual)
		}
	}
}

func TestRoundRobinStrategy_NoAliveBackends(t *testing.T) {
	backends := []*backend.Backend{
		{Addr: "localhost:9001", Alive: 0},
		{Addr: "localhost:9002", Alive: 0},
	}
	strategy := NewRoundRobinStrategy(backends)

	actual := strategy.GetNext()
	if actual != "" {
		t.Errorf("Expected empty string, got %s", actual)
	}
}
