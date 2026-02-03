// Package client provides client management and rate limiting functionality.
package client

import "time"

// Client represents an API client with rate limiting configuration.
// Each client is identified by a unique ID and authenticated via an API key.
type Client struct {
	ID         string    `json:"client_id"`    // Unique client identifier
	Capacity   int       `json:"capacity"`     // Token bucket capacity (burst size)
	RatePerSec int       `json:"rate_per_sec"` // Rate of token refill per second
	APIKey     string    `json:"api_key"`      // API key for authentication
	CreatedAt  time.Time `json:"created_at"`   // Timestamp of client creation
}
