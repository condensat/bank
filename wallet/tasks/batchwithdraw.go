package tasks

import (
	"context"
	"time"

	"git.condensat.tech/bank/logger"
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
	}
}
