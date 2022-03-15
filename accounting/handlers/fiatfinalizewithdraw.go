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

func FiatFinalizeWithdraw(ctx context.Context, authInfo common.AuthInfo, id uint64) (common.FiatFinalizeWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatFinalizeWithdraw")
	var result common.FiatFinalizeWithdraw

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	log.Debugf("operation ID: %v", id)
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

	result.ID = id
	result.UserName = string(user.Name)
	result.IBAN = common.IBAN(sepa.IBAN)
	result.Currency = string(updated.CurrencyName)
	result.Amount = float64(*(updated.Amount))

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
			if common.WithOperatorAuth {
				err := ValidateOtp(ctx, request.AuthInfo)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			operation, err := FiatFinalizeWithdraw(ctx, request.AuthInfo, request.ID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatFinalizeWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("FiatFinalizeWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
