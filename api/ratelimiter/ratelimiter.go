package ratelimiter

import (
	"context"
	"fmt"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"

	"github.com/go-redis/redis_rate/v8"
)

const (
	MiddlewarePeerRequestPerSecondKey = "Key.MiddlewarePeerRequestPerSecond"
	OpenSessionPerMinuteKey           = "Key.OpenSessionPerMinute"
)

type RateLimitInfo struct {
	redis_rate.Limit
	KeyPrefix string
}

type RateLimiter struct {
	limit     redis_rate.Limit
	limiter   *redis_rate.Limiter
	keyPrefix string
}

func New(ctx context.Context, rateLimit RateLimitInfo) *RateLimiter {
	rdb := cache.ToRedis(appcontext.Cache(ctx))
	return &RateLimiter{
		limit:     rateLimit.Limit,
		limiter:   redis_rate.NewLimiter(rdb),
		keyPrefix: rateLimit.KeyPrefix,
	}
}

func (p *RateLimiter) Allowed(ctx context.Context, name string) bool {
	key := fmt.Sprintf("%s:%s", p.keyPrefix, name)
	res, err := p.limiter.Allow(key, &p.limit)
	if err != nil {
		return false
	}

	return res.Allowed
}
