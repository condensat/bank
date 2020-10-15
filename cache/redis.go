package cache

import (
	"context"
	"fmt"

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

func (r *Redis) RDB() RDB {
	return r.rdb
}
