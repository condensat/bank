package client

import (
	"context"
	"time"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func FiatWithdraw(ctx context.Context, userId, accountId uint64, amount float64, currency, iban, bic, sepaLabel string) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatWithdraw")

	// amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	if len(iban) == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	dstIban := common.IBAN(iban)

	request := common.FiatWithdraw{
		UserId: userId,
		Source: common.AccountEntry{
			OperationType:    string(model.OperationTypeFiatWithdraw),
			SynchroneousType: "sync",
			Timestamp:        time.Now(),

			Amount:     amount, // withdraw remove amount from account
			LockAmount: 0.0,    // no lock on withdraw
			Currency:   currency,
		},
		Destination: common.FiatSepaInfo{
			Label: sepaLabel,
			IBAN:  dstIban,
			BIC:   bic,
		},
	}

	if accountId != 0 {
		request.Source.AccountID = accountId
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
