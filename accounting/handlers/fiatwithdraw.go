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

const (
	withOperatorAuth = true
)

func addFiatWithdrawToList(ctx context.Context, operation common.FiatOperationInfo) (model.FiatOperationInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.addFiatWithdrawToList")

	db := appcontext.Database(ctx)
	if db == nil {
		return model.FiatOperationInfo{}, errors.New("Invalid Database")
	}

	// Update Status to "pending"
	if operation.Status != "unvalidated" {
		return model.FiatOperationInfo{}, errors.New("Invalid Status")
	}

	operation.Status = "pending"

	// transform our struct in a model
	toWrite := model.FiatOperationInfo{
		Label:  operation.Label,
		IBAN:   operation.IBAN,
		BIC:    operation.BIC,
		Type:   operation.Type,
		Status: operation.Status,
	}

	// Write to db
	result, err := database.AddFiatOperationInfo(db, toWrite)
	if err != nil {
		return model.FiatOperationInfo{}, errors.New("AddFiatOperationInfo failed")
	}

	log.WithFields(logrus.Fields{
		"Label":  result.Label,
		"Type":   result.Type,
		"Status": result.Status,
	}).Debug("AddFiatOperationInfo success")

	return result, nil
}

func FiatWithdraw(ctx context.Context, authInfo common.AuthInfo, withdraw common.AccountEntry) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatWithdraw")

	db := appcontext.Database(ctx)
	if db == nil {
		return common.AccountEntry{}, errors.New("Invalid Database")
	}

	if withOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return common.AccountEntry{}, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return common.AccountEntry{}, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return common.AccountEntry{}, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return common.AccountEntry{}, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return common.AccountEntry{}, errors.New("CheckTOTP failed")
		}
		if !valid {
			return common.AccountEntry{}, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return common.AccountEntry{}, errors.New("Wrong operator ID")
		}
	}

	result, err := AccountOperation(ctx, withdraw)
	if err != nil {
		return common.AccountEntry{}, errors.New("AccountOperation failed")
	}

	log.WithFields(logrus.Fields{
		"Operation":       result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Currency":        result.Currency,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
		"Label":           result.Label,
	}).Debug("FiatWithdraw success")

	return result, err
}

func OnFiatWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			operation, err := FiatWithdraw(ctx, request.AuthInfo, request.Source)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatWithdraw")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"AccountID": operation.AccountID,
			})

			log.Info("FiatWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
