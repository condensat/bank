package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"github.com/sirupsen/logrus"
)

func FiatCancelWithdraw(ctx context.Context, authInfo common.AuthInfo, fiatOperationInfoId uint64, comment string) (common.FiatCancelWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatCancelWithdraw")

	var result common.FiatCancelWithdraw

	if fiatOperationInfoId == 0 {
		return result, cache.ErrInternalError
	}

	log = log.WithField("fiatOperationInfoId", fiatOperationInfoId)

	request := common.FiatCancelWithdraw{
		AuthInfo:            authInfo,
		FiatOperationInfoID: fiatOperationInfoId,
		Comment:             comment,
	}

	err := messaging.RequestMessage(ctx, common.FiatCancelWithdrawSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return result, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"ID":       result.FiatOperationInfoID,
		"UserName": result.UserName,
		"IBAN":     result.IBAN,
		"Currency": result.Currency,
		"Amount":   result.Amount,
	}).Debug("FiatCancelWithdraw registered")

	return result, nil
}
