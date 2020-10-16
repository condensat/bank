// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/wallet/common"

	"github.com/sirupsen/logrus"
)

func CryptoAddressNewDeposit(ctx context.Context, chain string, accountID uint64) (common.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.CryptoAddressNewDeposit")

	request := common.CryptoAddress{
		Chain:     chain,
		AccountID: accountID,
	}

	var result common.CryptoAddress
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CryptoAddressNewDepositSubject, &request, &result)
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
