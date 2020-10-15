package cache

import (
	"github.com/go-redis/redis/v8"
)

func ToRedis(cache Cache) *redis.Client {
	if cache == nil {
		return nil
	}
	rdb := cache.RDB()
	return rdb.(*redis.Client)
}
