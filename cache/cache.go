package cache

import (
	"git.condensat.tech/bank"

	"github.com/go-redis/redis/v7"
)

func ToRedis(cache bank.Cache) *redis.Client {
	if cache == nil {
		return nil
	}
	rdb := cache.RDB()
	return rdb.(*redis.Client)
}
