package cache

import (
	"git.condensat.tech/bank"

	"github.com/go-redis/redis"
)

func ToRedis(cache bank.Cache) *redis.Client {
	rdb := cache.RDB()
	return rdb.(*redis.Client)
}
