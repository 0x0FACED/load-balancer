package limitter

import (
	"context"
	"database/sql"
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

type TokenBucketLimitter struct {
	buckets map[string]*Bucket
	db      *sql.DB
	cfg     Config

	refillCancel context.CancelFunc
	mu           sync.RWMutex
}

func NewTokenBucketLimitter(db *sql.DB, cfg Config) *TokenBucketLimitter {
	return &TokenBucketLimitter{
		buckets: make(map[string]*Bucket),
		db:      db,
		cfg:     cfg,
	}
}

func (l *TokenBucketLimitter) Allow(clientID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, exists := l.buckets[clientID]
	if !exists {
		bucket = &Bucket{
			Capacity:       10,
			Tokens:         10,
			RefillRate:     1,
			LastRefillTime: time.Now(),
		}
		l.buckets[clientID] = bucket
	}

	now := time.Now()
	elapsed := now.Sub(bucket.LastRefillTime)
	bucket.Tokens += int(elapsed.Seconds()) * bucket.RefillRate
	if bucket.Tokens > bucket.Capacity {
		bucket.Tokens = bucket.Capacity
	}
	bucket.LastRefillTime = now

	if bucket.Tokens > 0 {
		bucket.Tokens--
		return true
	}

	return false
}

func (l *TokenBucketLimitter) Reset(clientID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, exists := l.buckets[clientID]
	if exists {
		bucket.Tokens = bucket.Capacity
		bucket.LastRefillTime = time.Now()
	} else {
		bucket = &Bucket{
			Capacity:       10, // default capacity
			Tokens:         10, // default tokens
			RefillRate:     1,  // default refill rate
			LastRefillTime: time.Now(),
		}
		l.buckets[clientID] = bucket
	}
}

func (rl *TokenBucketLimitter) Stop() error {
	if rl.refillCancel != nil {
		rl.refillCancel()
	}

	if rl.db != nil {
		if err := rl.db.Close(); err != nil {
			return err
		}
	}
	
	return nil
}

func (rl *TokenBucketLimitter) StartRefillJob(ctx context.Context) {
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

func (rl *TokenBucketLimitter) refillBuckets() {
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
