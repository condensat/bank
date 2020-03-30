package handlers

import (
	"context"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/accounting/internal"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountHistory(ctx context.Context, accountID uint64, from, to time.Time) ([]common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountHistory")
	var result []common.AccountEntry

	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
		"From":      from,
		"To":        to,
	})

	// Acquire Lock
	lock, err := internal.LockAccount(ctx, accountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return result, internal.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db bank.Database) error {
		account, err := database.GetAccountByID(db, model.AccountID(accountID))
		if err != nil {
			return err
		}

		operations, err := database.GeAccountHistoryRange(db, account.ID, from, to)
		if err != nil {
			return err
		}

		for _, op := range operations {
			if !op.IsValid() {
				return database.ErrInvalidAccountOperation
			}

			result = append(result, common.AccountEntry{
				AccountID: uint64(op.AccountID),
				Currency:  string(account.CurrencyName),

				OperationType:    string(op.OperationType),
				SynchroneousType: string(op.SynchroneousType),

				Timestamp: op.Timestamp,
				Label:     "N/A",
				Amount:    float64(*op.Amount),
				Balance:   float64(*op.Balance),

				LockAmount:  float64(*op.LockAmount),
				TotalLocked: float64(*op.TotalLocked),
			})
		}

		return nil
	})

	if err == nil {
		log.
			WithField("Count", len(result)).
			Debug("Account history retrieved")
	}

	return result, err
}

func OnAccountHistory(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountHistory")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountHistory
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
			})

			history, err := AccountHistory(ctx, request.AccountID, request.From, request.To)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to get AccountHistory")
				return nil, internal.ErrInternalError
			}

			// create & return response
			return &common.AccountHistory{
				AccountID: request.AccountID,
				From:      request.From,
				To:        request.To,

				History: history,
			}, nil
		})
}
