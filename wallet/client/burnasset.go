package client

import (
	"context"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/wallet/common"
)

// AssetBurn burns amount of asset by spending it to an unspendable output
func AssetBurn(ctx context.Context, chain string, issuerID uint64, assetToBurn string, amountToBurn float64) (common.BurnResponse, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.AssetBurn")

	var request common.BurnRequest

	request.Chain = chain
	request.IssuerID = issuerID

	request.Asset = assetToBurn
	request.Amount = amountToBurn

	var result common.BurnResponse
	err := messaging.RequestMessage(ctx, common.AssetBurnSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.BurnResponse{}, messaging.ErrRequestFailed
	}

	return result, nil
}
