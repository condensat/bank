package client

import (
	"context"
	"time"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func FiatDeposit(ctx context.Context, authInfo common.AuthInfo, userName string, amount float64, currency, label string) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatDeposit")

	if len(userName) == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	// amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	log = log.WithField("UserName", userName)

	request := common.FiatDeposit{
		AuthInfo: authInfo,
		UserName: userName,
		Destination: common.AccountEntry{
			OperationType:    "fiat_deposit",
			SynchroneousType: "sync",
			Timestamp:        time.Now(),

			Label: label,

			Amount:     amount,
			LockAmount: 0.0,
			Currency:   currency,
		},
	}

	var result common.AccountEntry
	err := messaging.RequestMessage(ctx, common.FiatDepositSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountEntry{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Operation":       result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Currency":        result.Currency,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
		"Label":           result.Label,
	}).Info("FiatDeposit registered")

	return result, nil
}
