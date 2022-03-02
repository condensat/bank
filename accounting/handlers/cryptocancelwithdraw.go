package handlers

import (
	"context"
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"github.com/sirupsen/logrus"
)

func CryptoCancelWithdraw(ctx context.Context, authInfo common.AuthInfo, id uint64, comment string) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CryptoCancelWithdraw")
	var result common.WithdrawInfo

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	result, err := CancelWithdraw(ctx, id, comment)
	if err != nil {
		return result, err
	}

	log.WithFields(logrus.Fields{
		"WithdrawID": result.WithdrawID,
		"AccountID":  result.AccountID,
		"Amount":     result.Amount,
		"Chain":      result.Chain,
		"Address":    result.PublicKey,
		"Comment":    comment,
	}).Info("Canceled withdraw")

	return result, nil
}

func OnCryptoCancelWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnCryptoWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CryptoCancelWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			var operatorID uint64
			if common.WithOperatorAuth {
				var err error
				operatorID, err = ValidateOtp(ctx, request.AuthInfo, common.CommandCryptoCancelWithdraw)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			operation, err := CryptoCancelWithdraw(ctx, request.AuthInfo, request.WithdrawID, request.Comment)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoCancelWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("CryptoCancelWithdraw succeeded")

			if common.WithOperatorAuth {
				// Update operator table
				err = UpdateOperatorTable(ctx, operatorID, operation.AccountID, 0)
				if err != nil {
					// not a fatal error, log an error and continue
					log.WithError(err).Error("UpdateOperatorTable failed")
				}
			}

			// create & return response
			return &operation, nil
		})
}
