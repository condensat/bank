package client

import (
	"context"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/wallet/common"

	"github.com/sirupsen/logrus"
)

func WalletList(ctx context.Context) ([]string, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.WalletList")

	var request common.WalletStatus
	var response common.WalletStatus
	err := messaging.RequestMessage(ctx, common.WalletListSubject, &request, &response)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return nil, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(response.Wallets),
	}).Debug("Wallet Info")

	var result []string
	for _, walletInfo := range response.Wallets {
		if len(walletInfo.Chain) == 0 {
			continue
		}
		result = append(result, walletInfo.Chain)
	}
	return result, nil
}
