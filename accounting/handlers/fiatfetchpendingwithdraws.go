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

func FiatFetchPendingWithdraw(ctx context.Context, authInfo common.AuthInfo) ([]common.FiatFetchPendingWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatFetchPendingWithdraw")

	db := appcontext.Database(ctx)
	if db == nil {
		return []common.FiatFetchPendingWithdraw{}, errors.New("Invalid Database")
	}

	if withOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return []common.FiatFetchPendingWithdraw{}, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return []common.FiatFetchPendingWithdraw{}, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return []common.FiatFetchPendingWithdraw{}, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return []common.FiatFetchPendingWithdraw{}, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return []common.FiatFetchPendingWithdraw{}, errors.New("CheckTOTP failed")
		}
		if !valid {
			return []common.FiatFetchPendingWithdraw{}, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return []common.FiatFetchPendingWithdraw{}, errors.New("Wrong operator ID")
		}
	}

	list, err := database.FetchFiatPendingWithdraw(db)
	if err != nil {
		return []common.FiatFetchPendingWithdraw{}, err
	}

	log.Debugf("Length of list: %v\n", len(list))
	result, err := convertFiatOperation(db, list)
	if err != nil {
		return []common.FiatFetchPendingWithdraw{}, err
	}

	// log.WithFields(logrus.Fields{
	// 	"Currency": result.Currency,
	// 	"Amount":   result.Amount,
	// 	"UserName": result.UserName,
	// }).Debug("FiatFetchPendingWithdraw success")
	log.Debug("FiatFetchPendingWithdraw success")

	return result, err
}

func convertFiatOperation(db bank.Database, list []model.FiatOperationInfo) ([]common.FiatFetchPendingWithdraw, error) {
	var result []common.FiatFetchPendingWithdraw
	for _, withdraw := range list {
		// look up the sepa info in db
		sepaInfo, err := database.GetSepaByID(db, withdraw.SepaInfoID)
		if err != nil {
			return []common.FiatFetchPendingWithdraw{}, err
		}

		// get the username from userID
		user, err := database.FindUserById(db, withdraw.UserID)
		if err != nil {
			return []common.FiatFetchPendingWithdraw{}, err
		}

		// append the fetchPendingWithdraw to list
		result = append(result, common.FiatFetchPendingWithdraw{
			ID:       uint64(withdraw.ID),
			UserName: string(user.Name),
			Currency: string(withdraw.CurrencyName),
			Amount:   float64(*withdraw.Amount),
			IBAN:     string(sepaInfo.IBAN),
			BIC:      string(sepaInfo.BIC),
		})
	}

	return result, nil
}

func OnFiatFetchPendingWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatFetchPendingWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AuthInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			list, err := FiatFetchPendingWithdraw(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatFetchPendingWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("FiatFetchPendingWithdraw succeeded")

			log.Debugf("length of pending withdraws list: %v\n", len(list))

			// create & return response
			return &common.FiatFetchPendingWithdrawList{
				PendingWithdraws: list[:],
			}, nil
		})
}
