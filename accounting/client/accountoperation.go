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
