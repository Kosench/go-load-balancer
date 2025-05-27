package health

import (
	"load-balancer/internal/backend"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	// Mock backend server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create mock backend
	backend := &backend.Backend{Addr: server.Listener.Addr().String()}

	// Run health check
	checkBackend(backend)

	if !backend.IsAlive() {
		t.Error("Backend should be marked as alive")
	}
}

func TestHealthCheck_UnhealthyBackend(t *testing.T) {
	// Mock a backend server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	backend := &backend.Backend{Addr: server.Listener.Addr().String()}

	// Run health check
	checkBackend(backend)

	if backend.IsAlive() {
		t.Error("Backend should be marked as unhealthy")
	}
}
