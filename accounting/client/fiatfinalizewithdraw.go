package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func FiatFinalizeWithdraw(ctx context.Context, authInfo common.AuthInfo, id uint64) (common.FiatFinalizeWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatFinalizeWithdraw")

	if id == 0 {
		return common.FiatFinalizeWithdraw{}, cache.ErrInternalError
	}

	log = log.WithField("id", id)

	request := common.FiatFinalizeWithdraw{
		AuthInfo: authInfo,
		ID:       id,
	}

	var result common.FiatFinalizeWithdraw
	err := messaging.RequestMessage(ctx, common.FiatFinalizeWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.FiatFinalizeWithdraw{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"ID":       result.ID,
		"UserName": result.UserName,
		"IBAN":     result.IBAN,
		"Currency": result.Currency,
		"Amount":   result.Amount,
	}).Debug("FiatFinalizeWithdrawal registered")

	return result, nil
}
