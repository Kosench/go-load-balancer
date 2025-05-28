package client

import "time"

type Client struct {
	ID         string    `json:"client_id"`
	Capacity   int       `json:"capacity"`
	RatePerSec int       `json:"rate_per_sec"`
	APIKey     string    `json:"api_key"`
	CreatedAt  time.Time `json:"created_at"`
}
