package client

import (
	"context"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/wallet/common"

	"github.com/sirupsen/logrus"
)

func ListIssuances(ctx context.Context, chain string, issuerID uint64, asset string) (common.IssuanceList, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.ListIssuances")

	var request common.ListIssuancesRequest
	var response common.IssuanceList

	request.Chain = chain
	request.IssuerID = issuerID
	request.Asset = asset
	err := messaging.RequestMessage(ctx, common.AssetListIssuancesSubject, &request, &response)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.IssuanceList{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Issuer ID": response.IssuerID,
		"Count":     len(response.Issuances),
	}).Debug("Issuances info")

	return response, nil
}
