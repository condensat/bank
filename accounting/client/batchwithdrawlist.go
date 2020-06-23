package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
)

func BatchWithdrawList(ctx context.Context, network string) (common.BatchWithdraws, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.BatchWithdrawList")
	log = log.WithField("Network", network)

	request := common.BatchWithdraw{
		Network: network,
	}

	var result common.BatchWithdraws
	err := messaging.RequestMessage(ctx, common.BatchWithdrawListSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.BatchWithdraws{}, messaging.ErrRequestFailed
	}

	log.WithField("Count", len(result.Batches)).
		Debug("BatchWithdraw List")

	return result, nil
}
