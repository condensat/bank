package api

import (
	"context"
	"time"

	"git.condensat.tech/bank/networking/ratelimiter"

	"github.com/go-redis/redis_rate/v9"
)

var (
	DefaultOpenSessionPerMinute = ratelimiter.RateLimitInfo{
		Limit: redis_rate.Limit{
			Period: time.Minute,
			Rate:   10,
			Burst:  10,
		},
		KeyPrefix: "OpenSession",
	}
)

func RegisterOpenSessionRateLimiter(ctx context.Context, rateLimit ratelimiter.RateLimitInfo) context.Context {
	rateLimit.Burst = rateLimit.Rate // see rate_limite.PerMinute
	raterLimiter := ratelimiter.New(ctx, rateLimit)
	return context.WithValue(ctx, ratelimiter.OpenSessionPerMinuteKey, raterLimiter)
}
