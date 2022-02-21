package handlers

import (
	"context"
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"github.com/sirupsen/logrus"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func CancelWithdraw(ctx context.Context, withdrawID uint64, comment string) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CancelWithdraw")

	result := common.WithdrawInfo{}

	if withdrawID == 0 {
		return result, database.ErrInvalidWithdrawID
	}

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// Database Query
	err := db.Transaction(func(db bank.Database) error {
		wi, err := database.GetLastWithdrawInfo(db, model.WithdrawID(withdrawID))
		if err != nil {
			log.WithError(err).
				Error("GetLastWithdrawInfo failed")
			return err
		}
		if wi.Status != model.WithdrawStatusCreated {
			log.WithField("Status", wi.Status).
				Error("Withraw status is not created")
			return cache.ErrInternalError
		}

		// Get the rest of the info
		w, err := database.GetWithdraw(db, model.WithdrawID(withdrawID))
		if err != nil {
			log.WithError(err).
				Error("GetWithdraw failed")
			return err
		}

		wt, err := database.GetWithdrawTargetByWithdrawID(db, model.WithdrawID(withdrawID))
		if err != nil {
			log.WithError(err).
				Error("GetWithdrawTargetByWithdrawID")
			return err
		}

		data, _ := wt.OnChainData()

		chain := data.Chain
		publicKey := data.PublicKey

		// Add the comment about the cancel
		var withdrawData model.Data
		cancelComment := model.WithdrawInfoCancelComment{
			Comment: comment,
		}
		withdrawData, err = model.EncodeData(&cancelComment)

		// Add a new WithdrawInfo entry for Withdraw
		wi, err = database.AddWithdrawInfo(db, model.WithdrawID(withdrawID), model.WithdrawStatusCanceling, model.WithdrawInfoData(withdrawData))
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawInfo failed")
			return err
		}
		if wi.Status != model.WithdrawStatusCanceling {
			log.WithField("Status", wi.Status).
				Error("Withraw status is not canceling")
			return cache.ErrInternalError
		}

		result.WithdrawID = uint64(wi.WithdrawID)
		result.Status = string(wi.Status)
		result.Amount = float64(*w.Amount)
		result.AccountID = uint64(w.From)
		result.Chain = chain
		result.PublicKey = publicKey

		return nil
	})
	if err != nil {
		return common.WithdrawInfo{}, err
	}

	return result, nil
}

func OnCancelWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnCancelWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CryptoCancelWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := CancelWithdraw(ctx, request.WithdrawID, request.Comment)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"WithdrawID": request.WithdrawID,
					}).Errorf("Failed to CancelWithdraw")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}
