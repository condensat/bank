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
)

func ListBatchWithdrawReady(ctx context.Context, network string) (common.BatchWithdraws, error) {
	return ListBatchWithdrawWithStatus(ctx, network, "ready")
}

func ListBatchWithdrawProcessing(ctx context.Context, network string) (common.BatchWithdraws, error) {
	return ListBatchWithdrawWithStatus(ctx, network, "processing")
}

func ListBatchWithdrawWithStatus(ctx context.Context, network, status string) (common.BatchWithdraws, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.ListBatchWithdrawWithStatus")
	log = log.WithField("Network", network)

	request := common.BatchWithdraw{
		Network: network,
		Status:  status,
	}

	var result common.BatchWithdraws
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.BatchWithdrawListSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.BatchWithdraws{}, messaging.ErrRequestFailed
	}

	log.WithField("Count", len(result.Batches)).
		Debug("BatchWithdraw List")

	return result, nil
}
