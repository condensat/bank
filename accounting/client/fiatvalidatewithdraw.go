package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func FiatValidateWithdraw(ctx context.Context, authInfo common.AuthInfo, id []uint64) (common.FiatValidWithdrawList, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatValidateWithdraw")

	var result common.FiatValidWithdrawList
	if len(id) == 0 {
		return result, cache.ErrInternalError
	}

	request := common.FiatValidateWithdraw{
		AuthInfo: authInfo,
		ID:       id,
	}

	err := messaging.RequestMessage(ctx, common.FiatValidateWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return result, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Number of withdraws validated": len(result.ValidatedWithdraws),
	}).Debug("CryptoValidateWithdraw registered")

	return result, nil
}
