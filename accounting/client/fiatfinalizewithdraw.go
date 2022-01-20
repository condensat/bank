package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func FiatFinalizeWithdraw(ctx context.Context, authInfo common.AuthInfo, userName, iban string) (common.FiatFinalizeWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatFinalizeWithdraw")

	if len(userName) == 0 {
		return common.FiatFinalizeWithdraw{}, cache.ErrInternalError
	}

	if len(iban) == 0 {
		return common.FiatFinalizeWithdraw{}, cache.ErrInternalError
	}

	log = log.WithField("userName", userName)

	request := common.FiatFinalizeWithdraw{
		AuthInfo: authInfo,
		UserName: userName,
		IBAN:     iban,
	}

	var result common.FiatFinalizeWithdraw
	err := messaging.RequestMessage(ctx, common.FiatFinalizeWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.FiatFinalizeWithdraw{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"UserName": result.UserName,
		"IBAN":     result.IBAN,
		"Currency": result.Currency,
		"Amount":   result.Amount,
	}).Debug("FiatFinalizeWithdrawal registered")

	return result, nil
}
