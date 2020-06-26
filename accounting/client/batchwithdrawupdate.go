package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"github.com/sirupsen/logrus"
)

func BatchWithdrawUpdate(ctx context.Context, batchID uint64, status, txID string) (common.BatchStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.BatchWithdrawUpdate")
	log = log.WithField("BatchID", batchID)

	request := common.BatchUpdate{
		BatchStatus: common.BatchStatus{
			BatchID: batchID,
			Status:  status,
		},
		TxID: txID,
	}

	var result common.BatchStatus
	err := messaging.RequestMessage(ctx, common.BatchWithdrawUpdateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.BatchStatus{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Status": request.Status,
		"TxID":   request.TxID,
	}).Debug("Batch updated")

	return result, nil
}
