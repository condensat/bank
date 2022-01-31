package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/security/utils"
	"github.com/sirupsen/logrus"
)

func FiatFinalizeWithdraw(ctx context.Context, authInfo common.AuthInfo, id uint64) (common.FiatFinalizeWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatFinalizeWithdraw")

	var result common.FiatFinalizeWithdraw
	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	if withOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return result, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return result, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return result, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return result, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return result, errors.New("CheckTOTP failed")
		}
		if !valid {
			return result, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return result, errors.New("Wrong operator ID")
		}
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
