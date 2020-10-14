package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/database/query"

	"github.com/sirupsen/logrus"
)

func AccountList(ctx context.Context, userID uint64) ([]common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountList")
	var result []common.AccountInfo

	log = log.WithField("UserID", userID)

	// Acquire Lock
	lock, err := cache.LockUser(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock user")
		return result, cache.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db database.Context) error {
		accounts, err := query.GetAccountsByUserAndCurrencyAndName(db, model.UserID(userID), "*", "*")
		if err != nil {
			return err
		}

		for _, account := range accounts {

			account, err := txGetAccountInfo(db, account)
			if err != nil {
				return err
			}

			result = append(result, account)
		}

		return nil
	})

	if err == nil {
		log.WithField("Count", len(result)).
			Debug("User accounts retrieved")
	}

	return result, err
}

func OnAccountList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.UserAccounts
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"UserID": request.UserID,
			})

			accounts, err := AccountList(ctx, request.UserID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to list user accounts")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.UserAccounts{
				UserID:   request.UserID,
				Accounts: accounts[:],
			}, nil
		})
}
