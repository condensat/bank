package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"git.condensat.tech/bank/api/ratelimiter"
	"git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/networking"

	"github.com/go-redis/redis_rate/v9"
)

var (
	ErrRateLimit = errors.New("RateLimitReached")

	DefaultPeerRequestPerSecond = ratelimiter.RateLimitInfo{
		Limit: redis_rate.Limit{
			Period: time.Second,
			Rate:   100,
			Burst:  100,
		},
		KeyPrefix: "PeerRequest",
	}

	DefaultOpenSessionPerMinute = ratelimiter.RateLimitInfo{
		Limit: redis_rate.Limit{
			Period: time.Minute,
			Rate:   10,
			Burst:  10,
		},
		KeyPrefix: "OpenSession",
	}
)

func RegisterRateLimiter(ctx context.Context, rateLimit ratelimiter.RateLimitInfo) context.Context {
	rateLimit.Burst = rateLimit.Rate // see rate_limite.PerSecond
	raterLimiter := ratelimiter.New(ctx, rateLimit)
	return context.WithValue(ctx, ratelimiter.MiddlewarePeerRequestPerSecondKey, raterLimiter)
}

func RegisterOpenSessionRateLimiter(ctx context.Context, rateLimit ratelimiter.RateLimitInfo) context.Context {
	rateLimit.Burst = rateLimit.Rate // see rate_limite.PerMinute
	raterLimiter := ratelimiter.New(ctx, rateLimit)
	return context.WithValue(ctx, ratelimiter.OpenSessionPerMinuteKey, raterLimiter)
}

// MiddlewarePeerRateLimiter return StatusTooManyRequests if rate limite is reached
func MiddlewarePeerRateLimiter(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := r.Context()

	switch limiter := ctx.Value(ratelimiter.MiddlewarePeerRequestPerSecondKey).(type) {
	case *ratelimiter.RateLimiter:

		if !limiter.Allowed(ctx, networking.RequesterIP(r)) {
			log := logger.Logger(ctx).WithField("Method", "api.MiddlewarePeerRateLimiter")

			networking.AppendRequestLog(log, r).
				WithError(ErrRateLimit).
				Warning("Too many requests")

			http.Error(rw, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next(rw, r)

	default:
		log := logger.Logger(ctx).WithField("Method", "api.MiddlewarePeerRateLimiter")

		networking.AppendRequestLog(log, r).
			WithError(services.ErrServiceInternalError).
			Error("No limiter found")

		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
