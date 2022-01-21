package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
)

func FiatFetchPendingWithdraw(ctx context.Context, authInfo common.AuthInfo) (common.FiatFetchPendingWithdrawList, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatFetchPendingWithdraw")

	request := authInfo

	var result common.FiatFetchPendingWithdrawList
	err := messaging.RequestMessage(ctx, common.FiatFetchPendingWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.FiatFetchPendingWithdrawList{}, messaging.ErrRequestFailed
	}

	log.Debug("FiatFetchPendingWithdraw succeed")

	return result, nil
}
