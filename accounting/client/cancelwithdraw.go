package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
)

func CancelWithdraw(ctx context.Context, authInfo common.AuthInfo, targetID uint64, comment string) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CancelWithdraw")
	log = log.WithField("TargetID", targetID)

	if targetID == 0 {
		return common.WithdrawInfo{}, cache.ErrInternalError
	}

	request := common.CancelWithdraw{
		AuthInfo: authInfo,
		TargetID: targetID,
		Comment:  comment,
	}

	var result common.WithdrawInfo
	err := messaging.RequestMessage(ctx, common.CancelWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.WithdrawInfo{}, messaging.ErrRequestFailed
	}

	return result, nil
}
