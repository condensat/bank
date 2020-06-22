package accounting

import (
	"context"

	"git.condensat.tech/bank/logger"
	"github.com/sirupsen/logrus"
)

type ChainOutput struct {
	PublicKey string
	Amount    float64
}

func SentWalletBatchRequest(ctx context.Context, chain string, outputs []ChainOutput) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.SentWalletBatchRequest")

	log.WithFields(logrus.Fields{
		"Chain":   chain,
		"Outputs": outputs,
	}).Debug("Sending batch to wallet")
	return nil
}
