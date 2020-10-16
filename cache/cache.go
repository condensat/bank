// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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
