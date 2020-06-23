package tasks

import (
	"context"
	"errors"
	"time"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
)

var (
	ErrProcessingBatchWithdraw = errors.New("Error Processing BatchWithdraw")
)

func BatchWithdraw(ctx context.Context, epoch time.Time, chains []string) {
	processBatchWithdraw(ctx, epoch, chains)
}

func processBatchWithdraw(ctx context.Context, epoch time.Time, chains []string) {
	log := logger.Logger(ctx).WithField("Method", "tasks.processBatchWithdraw")
	log = log.WithField("Epoch", epoch)

	for _, chain := range chains {
		log = log.WithField("Chain", chain)
		log.Debugf("Process Batch Withdraw")

		err := processBatchWithdrawChain(ctx, chain)
		if err != nil {
			log.WithError(err).Error("Failed to processBatchWithdrawChain")
			continue
		}
	}
}

func processBatchWithdrawChain(ctx context.Context, chain string) error {
	log := logger.Logger(ctx).WithField("Method", "tasks.processBatchWithdrawChain")

	// Acquire Lock
	lock, err := cache.LockBatchNetwork(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock batchNetwork")
		return ErrProcessingBatchWithdraw
	}
	defer lock.Unlock()

	return nil
}
