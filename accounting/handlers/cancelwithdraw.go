package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/logger"
	"github.com/sirupsen/logrus"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/database/query"

	"git.condensat.tech/bank/messaging"
)

func CancelWithdraw(ctx context.Context, withdrawID uint64) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CancelWithdraw")

	if withdrawID == 0 {
		return common.WithdrawInfo{}, query.ErrInvalidWithdrawID
	}

	result := common.WithdrawInfo{
		WithdrawID: withdrawID,
	}
	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db database.Context) error {
		wi, err := query.GetLastWithdrawInfo(db, model.WithdrawID(withdrawID))
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

		wi, err = query.AddWithdrawInfo(db, model.WithdrawID(withdrawID), model.WithdrawStatusCanceling, "{}")
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

		result.Status = string(wi.Status)
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

	var request common.WithdrawInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := CancelWithdraw(ctx, request.WithdrawID)
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
