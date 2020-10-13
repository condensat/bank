package services

import (
	"context"
	"fmt"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/networking/ratelimiter"
)

func OpenSessionAllowed(ctx context.Context, userID uint64) bool {
	switch limiter := ctx.Value(ratelimiter.OpenSessionPerMinuteKey).(type) {
	case *ratelimiter.RateLimiter:

		return limiter.Allowed(ctx, fmt.Sprintf("UserID:%d", userID))

	default:
		logger.Logger(ctx).WithField("Method", "services.OpenSessionAllowed").
			Error("Failed to get OpenSession Limiter")
		return false
	}
}
