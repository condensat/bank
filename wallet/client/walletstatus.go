package client

import (
	"context"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/wallet/common"

	"github.com/sirupsen/logrus"
)

func WalletStatus(ctx context.Context) (common.WalletStatus, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.WalletStatus")

	var request common.WalletStatus
	var result common.WalletStatus
	err := messaging.RequestMessage(ctx, common.WalletStatusSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.WalletStatus{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(result.Wallets),
	}).Debug("Wallet Info")

	return result, nil
}
