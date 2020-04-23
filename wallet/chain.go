package wallet

import (
	"context"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
)

type ChainState struct {
	Chain  string
	Height uint64
}

type AddressInfo struct {
	Chain         string
	PublicAddress string
	Mined         uint64 // 0 unknown, 1 mempool, BlockHeight
}

func FetchChainsState(ctx context.Context, chains ...string) ([]ChainState, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.FetchChainsState")

	var result []ChainState
	for _, chain := range chains {
		state, err := fetchChainState(ctx, chain)
		if err != nil {
			continue
		}

		result = append(result, state)
	}

	log.
		WithField("Count", len(result)).
		Debug("Chains State Fetched")

	return result, nil
}

func fetchChainState(ctx context.Context, chain string) (ChainState, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.fetchChainState")

	log = log.WithField("Chain", chain)

	// Acquire Lock
	lock, err := cache.LockChain(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock chain")
		return ChainState{}, cache.ErrLockError
	}
	defer lock.Unlock()

	// Todo: RPC call to chain daemon

	return ChainState{
		Chain:  chain,
		Height: 42,
	}, nil
}
