package limiter

import (
	"context"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type Bucket struct {
	Capacity       int
	Tokens         int
	RefillRate     int
	LastRefillTime time.Time

	mu sync.Mutex
}

type TokenBucketLimiter struct {
	buckets map[string]*Bucket
	repo    Repository
	cfg     Config

	refillCancel context.CancelFunc
	mu           sync.RWMutex
}

func NewTokenBucketLimiter(repo Repository, cfg Config) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets: make(map[string]*Bucket),
		repo:    repo,
		cfg:     cfg,
	}
}

func (rl *TokenBucketLimiter) Allow(ctx context.Context, clientID string) bool {
	rl.mu.RLock()
	bucket, exists := rl.buckets[clientID]
	rl.mu.RUnlock()

	if !exists {
		// double-check locking
		rl.mu.Lock()
		bucket, exists = rl.buckets[clientID]
		if !exists {
			cl, err := rl.repo.Get(ctx, clientID)
			if err != nil {
				rl.mu.Unlock()
				return false
			}
			if cl == nil {
				bucket = &Bucket{
					Capacity:       rl.cfg.Capacity,
					RefillRate:     rl.cfg.Rate,
					Tokens:         1,
					LastRefillTime: time.Now(),
				}
				rl.buckets[clientID] = bucket
			} else {
				bucket = &Bucket{
					Capacity:       cl.Capacity,
					RefillRate:     cl.RefillRate,
					Tokens:         1,
					LastRefillTime: time.Now(),
				}
				rl.buckets[clientID] = bucket
			}
		}
		rl.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if bucket.Tokens > 0 {
		bucket.Tokens--
		return true
	}

	return false
}

func (rl *TokenBucketLimiter) Reset(clientID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if exists {
		bucket.Tokens = bucket.Capacity
		bucket.LastRefillTime = time.Now()
	} else {
		bucket = &Bucket{
			Capacity:       rl.cfg.Capacity, // default capacity
			Tokens:         rl.cfg.Capacity, // default tokens
			RefillRate:     rl.cfg.Rate,     // default refill rate
			LastRefillTime: time.Now(),
		}
		rl.buckets[clientID] = bucket
	}
}

func (rl *TokenBucketLimiter) Stop() error {
	if rl.refillCancel != nil {
		rl.refillCancel()
	}

	if rl.repo != nil {
		if err := rl.repo.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (rl *TokenBucketLimiter) StartRefillJob(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	rl.refillCancel = cancel

	go func() {
		ticker := time.NewTicker(time.Duration(rl.cfg.RefillIntrerval) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.refillBuckets()
			}
		}
	}()
}

func (rl *TokenBucketLimiter) refillBuckets() {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	for _, bucket := range rl.buckets {
		bucket.mu.Lock()

		elapsed := time.Since(bucket.LastRefillTime).Seconds()
		tokensToAdd := int(elapsed * float64(bucket.RefillRate))
		if tokensToAdd > 0 {
			bucket.Tokens = min(bucket.Capacity, bucket.Tokens+tokensToAdd)
			bucket.LastRefillTime = time.Now()
		}

		bucket.mu.Unlock()
	}
}
