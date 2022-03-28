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

func CryptoValidateWithdraw(ctx context.Context, id []uint64) (common.CryptoValidatedWithdrawList, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CryptoValidateWithdraw")
	var result common.CryptoValidatedWithdrawList

	// Get all the targets
	wt, err := GetTargetList(ctx, id, model.WithdrawTargetOnChain)
	if err != nil {
		return result, err
	}

	// Process all the withdraws
	if len(wt) > 0 {
		err := ProcessWithdraws(ctx, wt)
		if err != nil {
			return result, err
		}
	} else {
		log.Info("No valid withdraw to process")
	}

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// If we are here, all the targets in the list were successfully validated
	// Now we return relevant informations
	for _, target := range wt {
		log.WithFields(logrus.Fields{
			"WithdrawID": target.WithdrawID,
			"TargetID":   target.ID,
		})

		withdraw := common.CryptoWithdraw{TargetID: uint64(target.ID)}

		w, err := database.GetWithdraw(db, target.WithdrawID)
		// an error is very unlikely at this stage, but if it happens for some reason
		// we don't want to return an error since the withdraw has already been validated
		if err != nil {
			log.WithError(err).
				Error("Failed to GetWithdraw")
			// Still append the targetID to let the operator know it succeeded
			result.ValidatedWithdraws = append(result.ValidatedWithdraws, withdraw)
			continue
		}

		// get data
		data, err := target.OnChainData()
		if err != nil {
			log.WithError(err).
				Error("Failed to get OnChainData")
			result.ValidatedWithdraws = append(result.ValidatedWithdraws, withdraw)
			continue
		}

		// Get userName
		accountID := w.From

		accountInfo, err := database.GetAccountByID(db, accountID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetAccountByID")
			result.ValidatedWithdraws = append(result.ValidatedWithdraws, withdraw)
			continue
		}

		userInfo, err := database.FindUserById(db, accountInfo.UserID)
		if err != nil {
			log.WithError(err).
				Error("Failed to FindUserByID")
			result.ValidatedWithdraws = append(result.ValidatedWithdraws, withdraw)
			continue
		}

		withdraw = common.CryptoWithdraw{
			WithdrawID: uint64(target.WithdrawID),
			TargetID:   uint64(target.ID),
			UserName:   string(userInfo.Name),
			Address:    data.PublicKey,
			Currency:   string(accountInfo.CurrencyName),
			Amount:     float64(*w.Amount),
		}
		log.WithFields(logrus.Fields{
			"WithdrawID": withdraw.WithdrawID,
			"TargetID":   withdraw.TargetID,
			"Currency":   withdraw.Currency,
			"Amount":     withdraw.Amount,
			"UserName":   withdraw.UserName,
			"Address":    withdraw.Address,
		}).Debug("Processed withdraw")

		result.ValidatedWithdraws = append(result.ValidatedWithdraws, withdraw)

	}
	return result, nil
}

func OnCryptoValidateWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnCryptoWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CryptoValidateWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			if common.WithOperatorAuth {
				_, err := ValidateOtp(ctx, request.AuthInfo, common.CommandCryptoValidateWithdraw)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			operation, err := CryptoValidateWithdraw(ctx, request.ID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("CryptoWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
