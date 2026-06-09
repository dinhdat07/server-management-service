package ratelimit

import (
	"context"
	"errors"
	"fmt"

	redisrate "github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	limiter *redisrate.Limiter
}

func NewRedisLimiter(client redis.UniversalClient) (*RedisLimiter, error) {
	if client == nil {
		return nil, errors.New("redis client is nil")
	}

	return &RedisLimiter{
		limiter: redisrate.NewLimiter(client),
	}, nil
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, policy Policy) (*Result, error) {
	if key == "" {
		return nil, errors.New("rate limit key is empty")
	}

	if policy.Limit <= 0 {
		return nil, fmt.Errorf("invalid rate limit for policy %q: %d", policy.Name, policy.Limit)
	}

	if policy.Window <= 0 {
		return nil, fmt.Errorf("invalid rate limit window for policy %q: %s", policy.Name, policy.Window)
	}

	burst := policy.Burst
	if burst <= 0 {
		burst = policy.Limit
	}

	res, err := l.limiter.Allow(ctx, key, redisrate.Limit{
		Rate:   policy.Limit,
		Burst:  burst,
		Period: policy.Window,
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		Allowed:    res.Allowed > 0,
		Limit:      policy.Limit,
		Remaining:  res.Remaining,
		RetryAfter: res.RetryAfter,
		ResetAfter: res.ResetAfter,
	}, nil
}
