package cache

import (
	"context"

	"git.condensat.tech/bank/wallet/chain"
)

func UpdateRedisChain(ctx context.Context, chainsStates ...chain.ChainState) error {
	// Todo: store chains states into redis
	return nil
}
