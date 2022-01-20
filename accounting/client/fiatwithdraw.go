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

func FiatWithdraw(ctx context.Context, authInfo common.AuthInfo, userName string, amount float64, currency, bankLabel, iban, bic, userLabel string) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatWithdraw")

	if len(userName) == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	// amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	if len(iban) == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	// TODO: check that currency is fiat

	log = log.WithField("userName", userName)

	request := common.FiatWithdraw{
		AuthInfo: authInfo,
		UserName: userName,
		Source: common.AccountEntry{
			OperationType:    "withdraw",
			SynchroneousType: "sync",
			Timestamp:        time.Now(),

			Label: bankLabel,

			Amount:     -amount, // withdraw remove amount from account
			LockAmount: 0.0,     // no lock on withdraw
			Currency:   currency,
		},
		Destination: common.FiatOperationInfo{
			Label: userLabel,
			IBAN:  iban,
			BIC:   bic,
		},
	}

	var result common.AccountEntry
	err := messaging.RequestMessage(ctx, common.FiatWithdrawSubject, &request, &result)
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
	}).Debug("FiatWithdrawal registered")

	return result, nil
}
