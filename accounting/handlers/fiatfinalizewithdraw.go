package handlers

import (
	"context"
	"errors"

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

func FiatFinalizeWithdraw(ctx context.Context, id uint64) (common.FiatFinalizeWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatFinalizeWithdraw")
	var result common.FiatFinalizeWithdraw

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// Now we only need to update the status of the fiat Operation
	var updated model.FiatOperationInfo
	updated, err := database.FiatOperationFinalize(db, model.FiatOperationInfoID(id))
	if err != nil {
		return result, err
	}

	user, err := database.FindUserById(db, updated.UserID)
	if err != nil {
		return result, err
	}

	sepa, err := database.GetSepaByID(db, updated.SepaInfoID)
	if err != nil {
		return result, err
	}

	// Get the accountID
	accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, updated.UserID, updated.CurrencyName, "*")
	if err != nil {
		return result, err
	}

	account := accounts[0]

	result.ID = id
	result.UserName = string(user.Name)
	result.IBAN = common.IBAN(sepa.IBAN)
	result.Currency = string(updated.CurrencyName)
	result.Amount = float64(*(updated.Amount))
	result.AccountID = uint64(account.ID)

	log.WithFields(logrus.Fields{
		"Currency": result.Currency,
		"Amount":   result.Amount,
		"UserName": result.UserName,
	}).Debug("FiatFinalizeWithdraw success")

	return result, err
}

func OnFiatFinalizeWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatFinalizeWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatFinalizeWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			var operatorID uint64
			if common.WithOperatorAuth {
				var err error
				operatorID, err = ValidateOtp(ctx, request.AuthInfo, common.CommandFiatFinalizeWithdraw)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			operation, err := FiatFinalizeWithdraw(ctx, request.ID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatFinalizeWithdraw")
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

			log.Info("FiatFinalizeWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
