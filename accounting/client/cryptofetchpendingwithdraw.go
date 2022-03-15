package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
)

func CryptoFetchPendingWithdraw(ctx context.Context, authInfo common.AuthInfo) (common.CryptoFetchPendingWithdrawList, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CryptoFetchPendingWithdraw")

	request := authInfo

	var result common.CryptoFetchPendingWithdrawList
	err := messaging.RequestMessage(ctx, common.CryptoFetchPendingWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CryptoFetchPendingWithdrawList{}, messaging.ErrRequestFailed
	}

	log.Debug("CryptoFetchPendingWithdraw succeed")

	return result, nil
}
