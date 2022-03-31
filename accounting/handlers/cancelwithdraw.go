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

func CancelWithdraw(ctx context.Context, targetID uint64, comment string) (common.WithdrawInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CancelWithdraw")

	result := common.WithdrawInfo{}

	if targetID == 0 {
		err := errors.New("Invalid withdrawID")
		log.Error(err)
		return result, err
	}

	log.WithField("TargetID", targetID)

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// Database Query
	err := db.Transaction(func(db bank.Database) error {
		wt, err := database.GetWithdrawTarget(db, model.WithdrawTargetID(targetID))
		if err != nil {
			log.WithError(err).
				Error("GetWithdrawTarget")
			return err
		}

		wi, err := database.GetLastWithdrawInfo(db, wt.WithdrawID)
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
		w, err := database.GetWithdraw(db, wt.WithdrawID)
		if err != nil {
			log.WithError(err).
				Error("GetWithdraw failed")
			return err
		}

		// Add the comment about the cancel
		var withdrawData model.Data
		cancelComment := model.WithdrawInfoCancelComment{
			Comment: comment,
		}
		withdrawData, err = model.EncodeData(&cancelComment)

		switch wt.Type {
		case model.WithdrawTargetOnChain:
			data, err := wt.OnChainData()
			if err != nil {
				log.WithError(err).Error("OnChainData failed")
				return err
			}

			chain := data.Chain
			publicKey := data.PublicKey

			// Add a new WithdrawInfo entry for Withdraw
			wi, err = database.AddWithdrawInfo(db, w.ID, model.WithdrawStatusCanceling, model.WithdrawInfoData(withdrawData))
			if err != nil {
				log.WithError(err).
					Error("AddWithdrawInfo failed")
				return err
			}

			result.WithdrawID = uint64(wi.WithdrawID)
			result.Status = string(wi.Status)
			result.Amount = float64(*w.Amount)
			result.AccountID = uint64(w.From)
			result.Chain = chain
			result.PublicKey = publicKey
			result.Type = string(wt.Type)

			return nil
		case model.WithdrawTargetSepa:
			data, err := wt.SepaData()
			if err != nil {
				log.WithError(err).Error("SepaData failed")
				return err
			}

			iban := data.IBAN

			// Add a new WithdrawInfo entry for Withdraw
			wi, err = database.AddWithdrawInfo(db, w.ID, model.WithdrawStatusCanceling, model.WithdrawInfoData(withdrawData))
			if err != nil {
				log.WithError(err).
					Error("AddWithdrawInfo failed")
				return err
			}

			result.WithdrawID = uint64(wi.WithdrawID)
			result.Status = string(wi.Status)
			result.Amount = float64(*w.Amount)
			result.AccountID = uint64(w.From)
			result.IBAN = common.IBAN(iban)
			result.Type = string(wt.Type)

			return nil

		default:
			return errors.New("Not implemented")
		}
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

	var request common.CancelWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			var operatorID uint64
			if common.WithOperatorAuth {
				var err error
				operatorID, err = ValidateOtp(ctx, request.AuthInfo, common.CommandCancelWithdraw)
				if err != nil {
					log.WithError(err).Error("Authentication failed")
					return nil, cache.ErrInternalError
				}
			}

			response, err := CancelWithdraw(ctx, request.TargetID, request.Comment)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"WithdrawID": request.TargetID,
					}).Errorf("Failed to CancelWithdraw")
				return nil, cache.ErrInternalError
			}

			if common.WithOperatorAuth {
				// Update operator table
				err = UpdateOperatorTable(ctx, operatorID, response.AccountID, response.WithdrawID)
				if err != nil {
					// not a fatal error, log an error and continue
					log.WithError(err).Error("UpdateOperatorTable failed")
				}
			}

			// return response
			return &response, nil
		})
}

func OnUserCancelWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnCancelWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CancelWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := CancelWithdraw(ctx, request.TargetID, request.Comment)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"WithdrawID": request.TargetID,
					}).Errorf("Failed to CancelWithdraw")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}
