package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(clientID string) bool
}

type RateLimiterConfig struct {
	Capacity   float64
	RefillRate float64
}

type tokenBucket struct {
	capacity    float64
	refillRate  float64
	tokens      float64
	lastUpdated time.Time
	mu          sync.Mutex
}

type TokenBucketRateLimiter struct {
	config    RateLimiterConfig
	buckets   map[string]*tokenBucket
	bucketsMu sync.RWMutex
}

func NewTokenBucketRateLimiter(cfg RateLimiterConfig) *TokenBucketRateLimiter {
	rl := &TokenBucketRateLimiter{
		config:  cfg,
		buckets: make(map[string]*tokenBucket),
	}
	return rl
}

// Allow проверяет, можно ли пропустить запрос от клиента с clientID.
func (rl *TokenBucketRateLimiter) Allow(clientID string) bool {
	rl.bucketsMu.Lock()
	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &tokenBucket{
			capacity:    rl.config.Capacity,
			refillRate:  rl.config.RefillRate,
			tokens:      float64(rl.config.Capacity),
			lastUpdated: time.Now(),
		}
		rl.buckets[clientID] = bucket
	}
	rl.bucketsMu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdated).Seconds()
	// Пополнение токенов только при обращении
	refilledTokens := bucket.tokens + elapsed*float64(bucket.refillRate)
	if refilledTokens > float64(bucket.capacity) {
		refilledTokens = float64(bucket.capacity)
	}
	bucket.tokens = refilledTokens
	bucket.lastUpdated = now

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}
	return false
}
