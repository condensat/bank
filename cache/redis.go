package cache

import (
	"context"
	"fmt"

	"git.condensat.tech/bank"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	rdb *redis.Client
}

func NewRedis(ctx context.Context, options RedisOptions) *Redis {
	return &Redis{
		rdb: redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", options.HostName, options.Port),
		}),
	}
}

func (r *Redis) RDB() bank.RDB {
	return r.rdb
}
