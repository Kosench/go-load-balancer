package ratelimit

import (
	"context"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(clientID string) bool
	Start(ctx context.Context)
}

type Config struct {
	Capacity   int
	RefillRate int
}

type TokenBucket struct {
	capacity    int
	refillRate  int
	tokens      int
	lastUpdated time.Time
	mu          sync.Mutex
}

type RateLimiterImpl struct {
	config    Config
	buckets   map[string]*TokenBucket
	bucketsMu sync.RWMutex
}

func NewRateLimiter(cfg Config) *RateLimiterImpl {
	rl := &RateLimiterImpl{
		config:  cfg,
		buckets: make(map[string]*TokenBucket),
	}
	return rl
}

func (rl *RateLimiterImpl) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				log.Info().Msg("Stopping rate limiter")
				return
			case t := <-ticker.C:
				rl.bucketsMu.RLock()
				for clientID, bucket := range rl.buckets {
					bucket.mu.Lock()
					elapsed := t.Sub(bucket.lastUpdated).Seconds()
					oldTokens := bucket.tokens
					bucket.tokens = min(bucket.capacity, bucket.tokens+int(elapsed)*bucket.refillRate)
					bucket.lastUpdated = t
					if bucket.tokens > oldTokens {
						log.Trace().
							Str("client", clientID).
							Int("tokens", bucket.tokens).
							Msg("Tokens refilled")
					}
					bucket.mu.Unlock()
				}
				rl.bucketsMu.RUnlock()
			}
		}
	}()
}

func (rl *RateLimiterImpl) Allow(clientID string) bool {
	rl.bucketsMu.Lock()
	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &TokenBucket{
			capacity:    rl.config.Capacity,
			refillRate:  rl.config.RefillRate,
			tokens:      rl.config.Capacity,
			lastUpdated: time.Now(),
		}
		rl.buckets[clientID] = bucket
	}
	rl.bucketsMu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Пополняем токены на основе времени
	elapsed := time.Since(bucket.lastUpdated).Seconds()
	bucket.tokens = min(bucket.capacity, bucket.tokens+int(elapsed)*bucket.refillRate)
	bucket.lastUpdated = time.Now()

	// Проверяем наличие токена
	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}
	return false
}

func minFloat(a, b int) int {
	if a < b {
		return a
	}
	return b
}
