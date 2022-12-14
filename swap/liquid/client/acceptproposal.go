package client

import (
	"context"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"git.condensat.tech/bank/swap/liquid/common"

	"github.com/sirupsen/logrus"
)

func AcceptSwapProposal(ctx context.Context, swapID uint64, address common.ConfidentialAddress, payload common.Payload, feeRate float64) (common.SwapProposal, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.client.AcceptSwapProposal")

	if !payload.Valid() {
		return common.SwapProposal{}, common.ErrInvalidPayload
	}

	request := common.SwapProposal{
		SwapID:  swapID,
		Address: address,
		Payload: payload,
		FeeRate: feeRate,
	}

	var result common.SwapProposal
	err := messaging.RequestMessage(ctx, common.SwapAcceptProposalSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.SwapProposal{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SwapID": result.SwapID,
	}).Debug("Accept SwapProposal")

	return result, nil
}
