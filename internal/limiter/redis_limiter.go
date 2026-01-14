package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/0x0FACED/load-balancer/internal/client"
	"github.com/0x0FACED/zlog"
	"github.com/redis/go-redis/v9"
)

// Redis Tocken Bucket limiter is a distributed limiter
// with redis client as storage and token bucket algorithm
type RedisTokenBucketLimiter struct {
	cl            *redis.Client
	fallbackInMem *TokenBucketLimiter // fallback to local cache if redis is not responding
	logger        *zlog.ZerologLogger
	repo          Repository
	cfg           Config
}

func NewRedisTocketBucketLimiter(
	client *redis.Client,
	inMem *TokenBucketLimiter,
	log *zlog.ZerologLogger,
	cfg Config,
) *RedisTokenBucketLimiter {
	return &RedisTokenBucketLimiter{
		cl:            client,
		fallbackInMem: inMem,
		logger:        log,
		cfg:           cfg,
	}
}

func (rl *RedisTokenBucketLimiter) Allow(ctx context.Context, clientID string) bool {
	cfg, err := rl.clientConfig(ctx, clientID)
	if err != nil {
		// log err and return or fallback to anonymous client
		return false
	}

	// not added to db, use default settings
	if cfg == nil {
		cfg = &client.Client{}

		cfg.ID = clientID
		cfg.Capacity = rl.cfg.Capacity
		cfg.RefillRate = rl.cfg.Rate
	}

	if rl.check(ctx,
		clientID,
		rl.cfg.TTL,
		cfg,
	) {
		return true
	}

	return false
}

func (rl *RedisTokenBucketLimiter) clientConfig(ctx context.Context, clientID string) (*client.Client, error) {
	// if repo added to limiter
	if rl.repo != nil {
		config, err := rl.repo.Get(ctx, clientID)
		if err != nil {
			return nil, err // TODO: handle err
		}

		return config, nil
	}

	// nil, nil - no error and repo is not initialized
	return nil, nil
}

// Lua script for getting client tokens by clientID, refilling tokens
// and calculating new tokens amount and returning 1 if Allowed, 0 otherwise.
//
// Its more perfect solution, because we dont need to iterate over ALL clients every tick.
// We just refill client tokens if he does request.
// All logic in 1 lua script.
var script string = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = tonumber(ARGV[4])

-- get curr state
local tokens = redis.call('HGET', key, 'tokens')
local last_refill = redis.call('HGET', key, 'last_refill')

-- init new client if there are no tokens
if tokens == false then
	redis.call('HSET', key, 'tokens', capacity - 1)
	redis.call('HSET', key, 'last_refill', now)
	redis.call('EXPIRE', key, ttl)
	return 1 -- allowed
end

tokens = tonumber(tokens)
last_refill = tonumber(last_refill)

-- refill tokens using time
local elapsed = now - last_refill
local tokens_to_add = math.floor(elapsed * refill_rate)

if tokens_to_add > 0 then
	tokens = math.min(capacity, tokens + tokens_to_add)
	redis.call('HSET', key, 'last_refill', now)
end

-- check and take client tokens
if tokens > 0 then
	tokens = tokens - 1
	redis.call('HSET', key, 'tokens', tokens)
	redis.call('EXPIRE', key, ttl)  -- обновляем TTL
	return 1 -- allowed
else
	return 0 -- rejected
end
`

func (rl *RedisTokenBucketLimiter) check(ctx context.Context, clientID string, ttl int, cfg *client.Client) bool {
	key := fmt.Sprintf("rate_limit:%s", clientID)
	now := time.Now().Unix()

	result, err := rl.cl.Eval(
		ctx,
		script,
		[]string{key},
		cfg.Capacity,
		cfg.RefillRate,
		now,
		ttl,
	).Int()
	if err != nil {
		// log err and return
		return false
	}

	// allow
	if result == 1 {
		return true
	}

	return false
}

func (rl *RedisTokenBucketLimiter) Reset(clientID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("rate_limit:%s", clientID)

	err := rl.cl.Del(ctx, key).Err()
	if err != nil {
		// log err
	}
}

func (rl *RedisTokenBucketLimiter) StartRefillJob(ctx context.Context) {
	// no need to impl really
}

func (rl *RedisTokenBucketLimiter) Stop() error {
	if rl.cl != nil {
		// log err and return
		return rl.cl.Close()
	}

	return nil
}
