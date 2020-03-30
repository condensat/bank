package internal

import (
	"context"
	"errors"
)

const (
	RedisLockerKey = "Key.RedisLockerKey"
)

var (
	ErrInternalError = errors.New("InternalError")
)

func RedisMutexContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, RedisLockerKey, NewRedisMutex(ctx))
}

func RedisMutexFromContext(ctx context.Context) Mutex {
	switch redisMutex := ctx.Value(RedisLockerKey).(type) {
	case *RedisMutex:
		return redisMutex

	case Mutex:
		return redisMutex

	default:
		return nil
	}
}
