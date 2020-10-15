package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
)

func CancelWithdraw(ctx context.Context, withdrawID uint64) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CancelWithdraw")
	log = log.WithField("UserID", withdrawID)

	if withdrawID == 0 {
		return common.WithdrawInfo{}, cache.ErrInternalError
	}

	var result common.WithdrawInfo
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CancelWithdrawSubject, &common.WithdrawInfo{WithdrawID: withdrawID}, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.WithdrawInfo{}, messaging.ErrRequestFailed
	}

	return result, nil
}
