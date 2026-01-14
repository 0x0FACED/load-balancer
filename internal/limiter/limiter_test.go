package limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/0x0FACED/load-balancer/internal/limiter"
	"github.com/0x0FACED/load-balancer/internal/limiter/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAllow_NewClient(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	cfg := limiter.Config{Capacity: 3, Rate: 1, RefillIntrerval: 100}
	lim := limiter.NewTokenBucketLimiter(mockRepo, cfg)

	mockRepo.On("Get", mock.Anything, "user1").Return(nil, nil)

	allowed := lim.Allow(context.Background(), "user1")
	assert.True(t, allowed, "first token should be allowed")
}

func TestAllow_NotEnoughTokens(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	cfg := limiter.Config{Capacity: 2, Rate: 1, RefillIntrerval: 100}
	lim := limiter.NewTokenBucketLimiter(mockRepo, cfg)

	mockRepo.On("Get", mock.Anything, "user2").Return(nil, nil)

	assert.True(t, lim.Allow(context.Background(), "user2"))
	assert.False(t, lim.Allow(context.Background(), "user2"), "should not allow more than capacity")
}

func TestReset(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	cfg := limiter.Config{Capacity: 2, Rate: 1, RefillIntrerval: 100}
	lim := limiter.NewTokenBucketLimiter(mockRepo, cfg)

	mockRepo.On("Get", mock.Anything, "user3").Return(nil, nil)

	lim.Allow(context.Background(), "user3")
	lim.Allow(context.Background(), "user3")
	lim.Reset("user3")

	assert.True(t, lim.Allow(context.Background(), "user3"), "token should be available after reset")
}

func TestRefillJob(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	cfg := limiter.Config{Capacity: 2, Rate: 2, RefillIntrerval: 100}
	lim := limiter.NewTokenBucketLimiter(mockRepo, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	lim.StartRefillJob(ctx)
	defer lim.Stop()

	mockRepo.On("Get", mock.Anything, "user4").Return(nil, nil)
	mockRepo.On("Close").Return(nil)

	lim.Allow(ctx, "user4")
	lim.Allow(ctx, "user4") // false

	time.Sleep(1000 * time.Millisecond)

	assert.True(t, lim.Allow(context.Background(), "user4"), "token should be refilled")
	cancel()
}
