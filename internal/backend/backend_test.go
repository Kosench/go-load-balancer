package backend

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestBackend_IsAlive(t *testing.T) {
	b := &Backend{}
	b.SetAlive(true)
	if !b.IsAlive() {
		t.Error("Expected backend to be alive")
	}

	b.SetAlive(false)
	if b.IsAlive() {
		t.Error("Expected backend to be not alive")
	}
}

func TestStartBackend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	backend := &Backend{Addr: "localhost:9999"}
	StartBackend(ctx, []*Backend{backend})

	// Allow time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Check if the server responds
	resp, err := http.Get("http://localhost:9999/")
	if err != nil {
		t.Fatalf("Failed to connect to backend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}
