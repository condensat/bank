package client

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"git.condensat.tech/bank/swap/liquid/common"

	"github.com/sirupsen/logrus"
)

func InfoSwapProposal(ctx context.Context, swapID uint64, payload common.Payload) (common.SwapProposal, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.client.SwapInfo")

	if !payload.Valid() {
		return common.SwapProposal{}, common.ErrInvalidPayload
	}

	request := common.SwapProposal{
		SwapID:  swapID,
		Payload: payload,
	}

	var result common.SwapProposal
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.SwapInfoProposalSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.SwapProposal{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SwapID": result.SwapID,
	}).Debug("Swap Info")

	return result, nil
}
