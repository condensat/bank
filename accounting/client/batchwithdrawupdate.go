// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"github.com/sirupsen/logrus"
)

func BatchWithdrawUpdate(ctx context.Context, batchID uint64, status, txID string) (common.BatchStatus, error) {
	return BatchWithdrawUpdateWithHeight(ctx, batchID, status, txID, 0)
}

func BatchWithdrawUpdateWithHeight(ctx context.Context, batchID uint64, status, txID string, height int) (common.BatchStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.BatchWithdrawUpdateWithHeight")
	log = log.WithField("BatchID", batchID)

	if height < 0 {
		height = 0
	}

	request := common.BatchUpdate{
		BatchStatus: common.BatchStatus{
			BatchID: batchID,
			Status:  status,
		},
		TxID:   txID,
		Height: height,
	}

	var result common.BatchStatus
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.BatchWithdrawUpdateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.BatchStatus{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"BatchID": request.BatchID,
		"Status":  request.Status,
		"TxID":    request.TxID,
		"Height":  request.Height,
	}).Debug("Batch updated")

	return result, nil
}
