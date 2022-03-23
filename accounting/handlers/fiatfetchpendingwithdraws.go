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

	var result []common.FiatFetchPendingWithdraw

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// Fetch the withdraws target
	wt, err := database.GetLastWithdrawTargetByStatus(db, model.WithdrawStatusCreated)
	if err != nil {
		return result, err
	}

	var targets []model.WithdrawTarget
	// Keep only the `sepa` type in the list
	for _, target := range wt {
		if target.Type == model.WithdrawTargetSepa {
			targets = append(targets, target)
		}
	}

	// with withdraws ID, we can fetch Withdraws
	for _, target := range targets {
		// get withdraw
		w, err := database.GetWithdraw(db, target.WithdrawID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetWithdraw")
			return result, err
		}
		// Get withdraw info history
		history, err := database.GetWithdrawHistory(db, target.WithdrawID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetWithdrawHistory")
			return result, errors.New("error")
		}
		// skip processed withdraw
		if len(history) != 1 || history[0].Status != model.WithdrawStatusCreated {
			log.Warn("Withdraw status is not created")
			continue
		}

		// get data
		data, err := target.SepaData()
		if err != nil {
			log.WithError(err).
				Error("Failed to get SepaData")
			return result, errors.New("error")
		}

		// Get userName
		accountID := w.From

		accountInfo, err := database.GetAccountByID(db, accountID)
		if err != nil {
			return result, err
		}

		userInfo, err := database.FindUserById(db, accountInfo.UserID)

		// append the fetchPendingWithdraw to list
		result = append(result, common.FiatFetchPendingWithdraw{
			ID:       uint64(w.ID),
			UserName: string(userInfo.Name),
			Currency: string(accountInfo.CurrencyName),
			Amount:   float64(*w.Amount),
			IBAN:     string(data.IBAN),
			BIC:      string(data.BIC),
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
