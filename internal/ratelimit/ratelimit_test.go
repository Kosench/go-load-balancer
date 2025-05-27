package ratelimit

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(Config{
		Capacity:   2,
		RefillRate: 1,
	})

	clientID := "clientID"

	if !limiter.Allow(clientID) {
		t.Error("First request should be allowed")
	}
	if !limiter.Allow(clientID) {
		t.Error("Second request should be allowed")
	}

	// Third request should be denied
	if limiter.Allow(clientID) {
		t.Error("Third request should be denied")
	}

	// Wait for refill and try again
	time.Sleep(1 * time.Second)
	if !limiter.Allow(clientID) {
		t.Error("Request after refill should be allowed")
	}

}
