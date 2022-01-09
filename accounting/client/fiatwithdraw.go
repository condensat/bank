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

func FiatWithdraw(ctx context.Context, authInfo common.AuthInfo, accountID, referenceID uint64, amount float64, currency string, label string, iban string, bic string) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.FiatWithdraw")

	if accountID == 0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	// amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, cache.ErrInternalError
	}

	// TODO: check that currency is fiat

	log = log.WithField("AccountID", accountID)

	request := common.FiatWithdraw{
		AuthInfo: authInfo,
		Source: common.AccountEntry{
			AccountID: accountID,

			ReferenceID:      referenceID,
			OperationType:    "withdraw",
			SynchroneousType: "sync",
			Timestamp:        time.Now(),

			Label: label,

			Amount:     -amount, // withdraw remove amount from account
			LockAmount: 0.0,     // no lock on withdraw
			Currency:   currency,
		},
		Destination: common.FiatOperationInfo{
			Label:  label,
			IBAN:   iban,
			BIC:    bic,
			Type:   model.OperationTypeWithdraw,
			Status: "unvalidated",
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
