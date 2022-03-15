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

func FiatCancelWithdraw(ctx context.Context, authInfo common.AuthInfo, fiatOperationInfoId uint64, comment string) (common.FiatCancelWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatCancelWithdraw")
	var result common.FiatCancelWithdraw

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
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
			if common.WithOperatorAuth {
				err := ValidateOtp(ctx, request.AuthInfo)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
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
