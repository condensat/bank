package client

import (
	"context"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/wallet/common"
)

// AssetReissuance reissues an asset if provided with a token input
func AssetReissuance(ctx context.Context, chain string, issuerID uint64, assetID, assetAddress string, assetAmount float64) (common.ReissuanceResponse, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.assetReissuance")

	var request common.ReissuanceRequest

	request.Chain = chain
	request.IssuerID = issuerID

	request.AssetID = assetID
	request.AssetPublicAddress = assetAddress
	request.AssetIssuedAmount = assetAmount

	var result common.ReissuanceResponse
	err := messaging.RequestMessage(ctx, common.AssetReissuanceSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.ReissuanceResponse{}, messaging.ErrRequestFailed
	}

	return result, nil
}
