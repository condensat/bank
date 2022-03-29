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

func FiatFetchPendingWithdraw(ctx context.Context) ([]common.FiatFetchPendingWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatFetchPendingWithdraw")

	db := appcontext.Database(ctx)
	if db == nil {
		return []common.FiatFetchPendingWithdraw{}, errors.New("Invalid Database")
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
			if common.WithOperatorAuth {
				_, err := ValidateOtp(ctx, request, common.CommandFiatFetchPendingWithdraw)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			list, err := FiatFetchPendingWithdraw(ctx)
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
