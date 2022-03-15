package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CryptoCancelWithdraw(ctx context.Context, authInfo common.AuthInfo, id uint64, comment string) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CryptoCancelWithdraw")

	var result common.WithdrawInfo

	if id == 0 {
		return result, cache.ErrInternalError
	}

	request := common.CryptoCancelWithdraw{
		AuthInfo:   authInfo,
		WithdrawID: id,
		Comment:    comment,
	}

	err := messaging.RequestMessage(ctx, common.CryptoCancelWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return result, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"WithdrawID": result.WithdrawID,
	}).Info("CryptoCancelWithdraw registered")

	return result, nil
}
