// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"context"

	"git.condensat.tech/bank/wallet/chain"
)

func UpdateRedisChain(ctx context.Context, chainsStates ...chain.ChainState) error {
	// Todo: store chains states into redis
	return nil
}
