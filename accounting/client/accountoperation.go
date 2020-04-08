package client

import (
	"context"
	"time"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/accounting/internal"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountDeposit(ctx context.Context, accountID, referenceID uint64, amount float64, label string) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountDeposit")

	if accountID == 0 {
		return common.AccountEntry{}, internal.ErrInternalError
	}

	// Deposit amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, internal.ErrInternalError
	}

	log = log.WithField("AccountID", accountID)

	request := common.AccountEntry{
		AccountID: accountID,

		ReferenceID:      referenceID,
		OperationType:    "deposit",
		SynchroneousType: "sync",
		Timestamp:        time.Now(),

		Label: label,

		Amount:     amount,
		LockAmount: 0.0, // no lock on deposit
	}

	var result common.AccountEntry
	err := messaging.RequestMessage(ctx, common.AccountOperationSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountEntry{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"OperationID":     result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
	}).Debug("Account amount")

	return result, nil
}

func AccountWithdraw(ctx context.Context, accountID, referenceID uint64, amount float64, label string) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountWithdraw")

	if accountID == 0 {
		return common.AccountEntry{}, internal.ErrInternalError
	}

	// Deposit amount must be positive
	if amount <= 0.0 {
		return common.AccountEntry{}, internal.ErrInternalError
	}

	log = log.WithField("AccountID", accountID)

	request := common.AccountEntry{
		AccountID: accountID,

		ReferenceID:      referenceID,
		OperationType:    "withdraw",
		SynchroneousType: "sync",
		Timestamp:        time.Now(),

		Label: label,

		Amount:     -amount, // withdraw remove amount from account
		LockAmount: 0.0,     // no lock on withdraw
	}

	var result common.AccountEntry
	err := messaging.RequestMessage(ctx, common.AccountOperationSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountEntry{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"OperationID":     result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
	}).Debug("Account Withdraw")

	return result, nil
}

func AccountTransfert(ctx context.Context, srcAccountID, dstAccountID, referenceID uint64, currency string, amount float64, label string) (common.AccountTransfert, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountTransfert")

	if srcAccountID == 0 || dstAccountID == 0 {
		return common.AccountTransfert{}, internal.ErrInternalError
	}
	if srcAccountID == dstAccountID {
		return common.AccountTransfert{}, internal.ErrInternalError
	}

	// currency must be valid
	if len(currency) == 0 {
		return common.AccountTransfert{}, internal.ErrInternalError
	}

	// deposit amount must be positive
	if amount <= 0.0 {
		return common.AccountTransfert{}, internal.ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"SrcAccountID": srcAccountID,
		"DstAccountID": dstAccountID,

		"Amount":   amount,
		"Currency": currency,
	})

	request := common.AccountTransfert{
		Source: common.AccountEntry{
			AccountID: srcAccountID,
			Currency:  currency,
		},
		Destination: common.AccountEntry{
			AccountID: dstAccountID,

			OperationType:    "transfert",
			SynchroneousType: "sync",
			ReferenceID:      referenceID,

			Timestamp: time.Now(),
			Amount:    amount,

			Label: label,

			LockAmount: 0.0, // no lock on sync account transfert
			Currency:   currency,
		},
	}

	var result common.AccountTransfert
	err := messaging.RequestMessage(ctx, common.AccountTransfertSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountTransfert{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SrcID":      result.Source.OperationID,
		"SrcPrevID":  result.Source.OperationPrevID,
		"SrcBalance": result.Source.Balance,

		"DstID":      result.Destination.OperationID,
		"DstPrevID":  result.Destination.OperationPrevID,
		"DstBalance": result.Destination.Balance,
	}).Debug("Account amount")

	return result, nil
}
