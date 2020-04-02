package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/accounting/internal"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/messaging"
	"github.com/sirupsen/logrus"

	"git.condensat.tech/bank/logger"
)

func AccountList(ctx context.Context, userID uint64) ([]common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountList")
	var result []common.AccountInfo

	log = log.WithField("UserID", userID)

	// Acquire Lock
	lock, err := internal.LockUser(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock user")
		return result, internal.ErrLockError
	}
	defer lock.Unlock()

	// Database Query
	db := appcontext.Database(ctx)
	err = db.Transaction(func(db bank.Database) error {
		accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, model.UserID(userID), "*", "*")
		if err != nil {
			return err
		}

		for _, account := range accounts {
			accountState, err := database.GetAccountStatusByAccountID(db, account.ID)
			if err != nil {
				return err
			}

			result = append(result, common.AccountInfo{
				AccountID: uint64(account.ID),
				Currency:  string(account.CurrencyName),
				Name:      string(account.Name),
				Status:    string(accountState.State),
			})
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
				return nil, internal.ErrInternalError
			}

			// create & return response
			return &common.UserAccounts{
				UserID:   request.UserID,
				Accounts: accounts[:],
			}, nil
		})
}
