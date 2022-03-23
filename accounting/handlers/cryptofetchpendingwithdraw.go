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

func CryptoFetchPendingWithdraw(ctx context.Context) ([]common.CryptoWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.FetchPendingWithdraws")

	var result []common.CryptoWithdraw

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
	// Keep only the `onchain` type in the list
	for _, target := range wt {
		if target.Type == model.WithdrawTargetOnChain {
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
		data, err := target.OnChainData()
		if err != nil {
			log.WithError(err).
				Error("Failed to get OnChainData")
			return result, errors.New("error")
		}

		// Get userName
		accountID := w.From

		accountInfo, err := database.GetAccountByID(db, accountID)
		if err != nil {
			return result, err
		}

		userInfo, err := database.FindUserById(db, accountInfo.UserID)

		userName := userInfo.Name

		result = append(result, common.CryptoWithdraw{
			WithdrawID: uint64(target.WithdrawID),
			TargetID:   uint64(target.ID),
			UserName:   string(userName),
			Address:    data.PublicKey,
			Amount:     float64(*w.Amount),
			Currency:   string(accountInfo.CurrencyName),
		})
	}

	return result, nil
}

func OnCryptoFetchPendingWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnCryptoFetchPendingWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AuthInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			if common.WithOperatorAuth {
				_, err := ValidateOtp(ctx, request, common.CommandCryptoFetchPendingWithdraw)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			list, err := CryptoFetchPendingWithdraw(ctx)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoFetchPendingWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("CryptoFetchPendingWithdraw succeeded")

			// create & return response
			return &common.CryptoFetchPendingWithdrawList{
				PendingWithdraws: list[:],
			}, nil
		})
}
