package backend

import (
	"context"
	"net/http"
	"sync"
	"testing"
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
	var wg sync.WaitGroup
	backend := &Backend{Addr: ":0"} // listen on random available port
	ready := make(chan struct{})
	StartBackend(ctx, []*Backend{backend}, &wg, ready)

	// Wait for backend to be ready
	<-ready

	// Check if the server responds
	resp, err := http.Get("http://" + backend.Addr + "/")
	if err != nil {
		cancel()
		wg.Wait()
		t.Fatalf("Failed to connect to backend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cancel()
		wg.Wait()
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	// Shutdown and wait
	cancel()
	wg.Wait()
}
