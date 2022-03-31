package handlers

import (
	"context"
	"errors"
	"fmt"
	"math"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"github.com/sirupsen/logrus"
)

func FiatDeposit(ctx context.Context, userName string, deposit common.AccountEntry) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatDeposit")

	var result common.AccountEntry
	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// Test for null or negative amount
	if deposit.Amount <= 0.0 {
		return result, errors.New("Null or negative amount")
	}

	// We'd better set an upper limit too, for now let's say it's MaxFloat
	if deposit.Amount >= math.MaxFloat64 {
		return result, errors.New("Absurdly high amount")
	}

	// Is currency fiat ?
	currencyInfo, err := database.GetCurrencyByName(db, model.CurrencyName(deposit.Currency))
	if err != nil {
		return result, err
	}

	var fiatType model.Int = 0
	if *currencyInfo.Type != fiatType {
		return result, errors.New("Currency is not fiat")
	}

	// is OperationType fiat_deposit?
	if deposit.OperationType != model.OperationTypeFiatDeposit.String() {
		return result, errors.New("OperationType is not fiat_deposit")
	}

	email := fmt.Sprintf("%s@condensat.tech", userName)

	user, err := database.FindUserByEmail(db, model.UserEmail(email))
	if err != nil {
		return result, err
	}

	if user.ID == 0 {
		return result, errors.New("userID can't be 0")
	}

	// Get AccountID with UserID
	account, err := database.GetAccountsByUserAndCurrencyAndName(db, user.ID, model.CurrencyName(deposit.Currency), model.AccountName("*"))
	if err != nil {
		return result, err
	}

	if len(account) == 0 {
		// Create a new account for this user and currency
		createdAccount, err := AccountCreate(ctx, uint64(user.ID), common.AccountInfo{
			Currency: common.CurrencyInfo{
				Name: deposit.Currency,
			},
		})
		if err != nil {
			return result, err
		}

		// Set new account to normal
		_, err = database.AddOrUpdateAccountState(db, model.AccountState{
			AccountID: model.AccountID(createdAccount.AccountID),
			State:     model.AccountStatusNormal,
		})
		if err != nil {
			log.WithError(err).Error("AddOrUpdateAccountState")
			return result, errors.New("Can't update new account state")
		}

		deposit.AccountID = uint64(createdAccount.AccountID)
	} else {
		deposit.AccountID = uint64(account[0].ID)
	}

	log = log.WithField("accountID", deposit.AccountID)

	// Set reference id as userID
	deposit.ReferenceID = uint64(user.ID)
	log = log.WithField("ReferenceID", deposit.ReferenceID)

	// Making the operation
	result, err = AccountOperation(ctx, deposit)
	if err != nil {
		log.WithError(err).Error("AccountOperation failed")
		return result, errors.New("AccountOperation failed")
	}

	result.Currency = deposit.Currency

	log.WithFields(logrus.Fields{
		"Operation":       result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Currency":        result.Currency,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
		"Label":           result.Label,
	}).Debug("FiatDeposit success")

	return result, err
}

func OnFiatDeposit(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatDeposit")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatDeposit
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			var operatorID uint64
			if common.WithOperatorAuth {
				var err error
				operatorID, err = ValidateOtp(ctx, request.AuthInfo, common.CommandFiatDeposit)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			operation, err := FiatDeposit(ctx, request.UserName, request.Destination)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatDeposit")
				return nil, cache.ErrInternalError
			}

			if common.WithOperatorAuth {
				// Update operator table
				err = UpdateOperatorTable(ctx, operatorID, operation.AccountID, operation.OperationID)
				if err != nil {
					// not a fatal error, log an error and continue
					log.WithError(err).Error("UpdateOperatorTable failed")
				}
			}

			log = log.WithFields(logrus.Fields{
				"AccountID": operation.AccountID,
			})

			log.Info("FiatDeposit succeeded")

			// create & return response
			return &operation, nil
		})
}
