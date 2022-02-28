package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountOperation(ctx context.Context, entry common.AccountEntry) (common.AccountEntry, error) {
	db := appcontext.Database(ctx)

	return AccountOperationWithDatabase(ctx, db, entry)
}

func AccountOperationWithDatabase(ctx context.Context, db bank.Database, entry common.AccountEntry) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountOperationWithDatabase")

	log = log.WithFields(logrus.Fields{
		"AccountID":        entry.AccountID,
		"Currency":         entry.Currency,
		"SynchroneousType": entry.SynchroneousType,
		"OperationType":    entry.OperationType,
		"ReferenceID":      entry.ReferenceID,
	})

	// Acquire Lock
	lock, err := cache.LockAccount(ctx, entry.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountEntry{}, cache.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	amount := model.Float(entry.Amount)
	lockAmount := model.Float(entry.LockAmount)

	// Balance & totalLocked ar computed by database later, must be valid for pre-check
	var totalLocked model.Float
	if totalLocked < lockAmount {
		totalLocked = lockAmount
	}
	var balance model.Float
	if balance < amount {
		balance = amount
	}
	if balance < lockAmount {
		balance = lockAmount
	}

	ops, err := database.TxAppendAccountOperationSlice(db,
		common.ConvertEntryToOperation(entry),
	)
	if err != nil {
		log.WithError(err).
			Error("Failed to TxAppendAccountOperationSlice")
		return common.AccountEntry{}, err
	}

	if len(ops) != 1 {
		return common.AccountEntry{}, err
	}

	op := ops[0]

	log.
		WithField("OperationID", op.ID).
		Trace("Account operation")

	return common.AccountEntry{
		OperationID: uint64(op.ID),

		AccountID:        uint64(op.AccountID),
		ReferenceID:      uint64(op.ReferenceID),
		OperationType:    string(op.OperationType),
		SynchroneousType: string(op.SynchroneousType),

		Timestamp: op.Timestamp,
		Label:     "N/A",
		Amount:    float64(*op.Amount),
		Balance:   float64(*op.Balance),

		LockAmount:  float64(*op.LockAmount),
		TotalLocked: float64(*op.TotalLocked),
	}, nil
}

func OnAccountOperation(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountOperation")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountEntry
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
			})

			response, err := AccountOperation(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AccountOperation")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &response, nil
		})
}
