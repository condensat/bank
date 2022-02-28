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

func FiatCancelWithdraw(ctx context.Context, authInfo common.AuthInfo, fiatOperationInfoId uint64, comment string) (common.FiatCancelWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatCancelWithdraw")
	var result common.FiatCancelWithdraw

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	if common.WithOperatorAuth {
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

	result, err := fiatCancelWithdraw(ctx, db, log, common.FiatCancelWithdraw{
		FiatOperationInfoID: fiatOperationInfoId,
		Comment:             comment,
	})
	if err != nil {
		return result, err
	}

	return result, nil
}

func fiatCancelWithdraw(ctx context.Context, db bank.Database, log *logrus.Entry, result common.FiatCancelWithdraw) (common.FiatCancelWithdraw, error) {
	// get FiatOperationInfo
	fiatOperationInfo, err := database.FindFiatOperationById(db, model.FiatOperationInfoID(result.FiatOperationInfoID))
	if err != nil {
		return common.FiatCancelWithdraw{}, err
	}

	// Check status
	if fiatOperationInfo.Status != model.FiatOperationStatusPending {
		log.WithField("fiatOperationInfo.Status", fiatOperationInfo.Status).Error("Not pending fiatOperation")
		return common.FiatCancelWithdraw{}, errors.New("Can't cancel a non pending operation")
	}

	// Get the accountID
	accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, fiatOperationInfo.UserID, fiatOperationInfo.CurrencyName, "*")
	if err != nil {
		return common.FiatCancelWithdraw{}, err
	}

	account := accounts[0]

	// Refund the user
	_, err = AccountOperation(ctx, common.AccountEntry{
		AccountID: uint64(account.ID),
		Currency:  string(account.CurrencyName),

		OperationType:    string(model.OperationTypeRefund),
		SynchroneousType: string(model.SynchroneousTypeSync),

		Amount:      float64(*fiatOperationInfo.Amount),
		ReferenceID: uint64(fiatOperationInfo.SepaInfoID),

		Timestamp: common.Timestamp(),
	})
	if err != nil {
		return common.FiatCancelWithdraw{}, errors.New("Failed to refund canceled withdraw")
	}

	var updated model.FiatOperationInfo
	updated, err = database.FiatOperationCancel(db, model.FiatOperationInfoID(result.FiatOperationInfoID))
	if err != nil {
		log.WithError(err).Error("FiatOperationCancel failed")
		return common.FiatCancelWithdraw{}, errors.New("Refund the cancel withdraw, but update of fiat Operation failed")
	}

	// TODO add a comment to comment table that points to the operation

	user, err := database.FindUserById(db, updated.UserID)
	if err != nil {
		return common.FiatCancelWithdraw{}, err
	}

	sepa, err := database.GetSepaByID(db, updated.SepaInfoID)
	if err != nil {
		return common.FiatCancelWithdraw{}, err
	}

	result.UserName = string(user.Name)
	result.IBAN = common.IBAN(sepa.IBAN)
	result.Currency = string(updated.CurrencyName)
	result.Amount = float64(*(updated.Amount))

	log.WithFields(logrus.Fields{
		"FiatOperationInfoID": result.FiatOperationInfoID,
		"UserName":            result.UserName,
		"Amount":              result.Amount,
		"IBAN":                result.IBAN,
		"Comment":             result.Comment,
	}).Info("Canceled withdraw")

	return result, nil
}

func OnFiatCancelWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatCancelWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatCancelWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			operation, err := FiatCancelWithdraw(ctx, request.AuthInfo, request.FiatOperationInfoID, request.Comment)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatCancelWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("FiatCancelWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
