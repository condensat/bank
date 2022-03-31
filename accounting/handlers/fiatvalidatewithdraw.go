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

func FiatValidateWithdraw(ctx context.Context, id []uint64) (common.FiatValidWithdrawList, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatValidWithdraw")
	var result common.FiatValidWithdrawList

	if len(id) == 0 {
		return result, errors.New("Empty list of withdraws")
	}

	// Get all the targets
	wt, err := GetTargetList(ctx, id, model.WithdrawTargetSepa)
	if err != nil {
		return result, err
	}

	// Now we can process all the withdraws
	if len(wt) > 0 {
		err := ProcessWithdraws(ctx, wt)
		if err != nil {
			return result, err
		}
	} else {
		log.Info("No valid withdraw to process")
		return result, errors.New("No valid withdrawIDs provided")
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

		withdraw := common.FiatValidWithdraw{TargetID: uint64(target.ID)}

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
		data, err := target.SepaData()
		if err != nil {
			log.WithError(err).
				Error("Failed to get SepaData")
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

		withdraw = common.FiatValidWithdraw{
			WithdrawID: uint64(target.WithdrawID),
			TargetID:   uint64(target.ID),
			UserName:   string(userInfo.Name),
			IBAN:       common.IBAN(data.IBAN),
			Currency:   string(accountInfo.CurrencyName),
			Amount:     float64(*w.Amount),
			AccountID:  uint64(accountID), // We need this only for registering operator's action
		}
		log.WithFields(logrus.Fields{
			"WithdrawID": withdraw.WithdrawID,
			"TargetID":   withdraw.TargetID,
			"Currency":   withdraw.Currency,
			"Amount":     withdraw.Amount,
			"UserName":   withdraw.UserName,
			"IBAN":       withdraw.IBAN,
		}).Debug("Processed withdraw")

		result.ValidatedWithdraws = append(result.ValidatedWithdraws, withdraw)
	}

	return result, nil
}

func OnFiatValidateWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatValidateWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatValidateWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			var operatorID uint64
			if common.WithOperatorAuth {
				var err error
				operatorID, err = ValidateOtp(ctx, request.AuthInfo, common.CommandFiatValidateWithdraw)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}
			operations, err := FiatValidateWithdraw(ctx, request.ID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatFinalizeWithdraw")
				return nil, cache.ErrInternalError
			}

			if common.WithOperatorAuth {
				for _, op := range operations.ValidatedWithdraws {
					// Update operator table
					err = UpdateOperatorTable(ctx, operatorID, op.AccountID, 0)
					if err != nil {
						// not a fatal error, log an error and continue
						log.WithError(err).Error("UpdateOperatorTable failed")
					}
				}
			}

			log.Info("FiatValidateWithdraw succeeded")

			// create & return response
			return &operations, nil
		})
}
