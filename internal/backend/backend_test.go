package backend

import (
	"context"
	"net/http"
	"sync"
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
	var wg sync.WaitGroup
	backend := &Backend{Addr: ":0"} // listen on random available port
	StartBackend(ctx, []*Backend{backend}, &wg)

	// Allow time for the server to start and Addr to be filled
	time.Sleep(200 * time.Millisecond)

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
