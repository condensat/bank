package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func BatchWithdrawUpdate(ctx context.Context, batchID uint64, status, txID string) (common.BatchStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.BatchWithdrawUpdate")

	// Database Query
	db := appcontext.Database(ctx)

	batchInfo, err := database.GetLastBatchInfo(db, model.BatchID(batchID))
	if err != nil {
		log.WithError(err).
			Error("Failed to GetLastBatchInfoByStatusAndNetwork")
		return common.BatchStatus{}, err
	}
	if batchInfo.Status != model.BatchStatusReady {
		log.WithError(err).
			Error("Previous batch status is not ready")
		return common.BatchStatus{}, err
	}
	if status != string(model.BatchStatusProcessing) {
		log.WithError(err).
			Error("Previous batch status is not processing")
		return common.BatchStatus{}, err
	}
	if len(txID) != 0 {
		log.Error("Invalid txID")
		return common.BatchStatus{}, database.ErrInvalidTransactionID
	}

	// change status to processing, with TxID
	data, err := model.EncodeData(&model.BatchInfoCryptoData{
		TxID: model.String(txID),
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to EncodeData")
		return common.BatchStatus{}, err
	}

	batchInfo, err = database.AddBatchInfo(db, batchInfo.BatchID, model.BatchStatusProcessing, batchInfo.Type, model.BatchInfoData(data))
	if err != nil {
		log.WithError(err).
			Error("Failed to AddBatchInfos")
		return common.BatchStatus{}, err
	}

	return common.BatchStatus{
		BatchID: batchID,
		Status:  string(batchInfo.Status),
	}, err
}

func OnBatchWithdrawUpdate(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnBatchWithdrawUpdate")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.BatchUpdate
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"BatchID": request.BatchID,
				"TxID":    request.TxID,
			})

			response, err := BatchWithdrawUpdate(ctx, request.BatchID, request.Status, request.TxID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to update batch withdraws")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &response, nil
		})
}
