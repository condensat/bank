package client

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/wallet/common"

	"github.com/sirupsen/logrus"
)

func CryptoAddressNextDeposit(ctx context.Context, chain string, accountID uint64) (common.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.CryptoAddressNextDeposit")

	request := common.CryptoAddress{
		Chain:     chain,
		AccountID: accountID,
	}

	var result common.CryptoAddress
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CryptoAddressNextDepositSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CryptoAddress{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"CryptoAddressID": result.CryptoAddressID,
		"Chain":           result.Chain,
		"AccountID":       result.AccountID,
		"PublicAddress":   result.PublicAddress,
	}).Debug("Next Deposit Address")

	return result, nil
}
