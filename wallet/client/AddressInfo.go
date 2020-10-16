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

func AddressInfo(ctx context.Context, chain, publicAddress string) (common.AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.client.AddressInfo")

	request := common.CryptoAddress{
		Chain:         chain,
		PublicAddress: publicAddress,
	}

	var result common.AddressInfo
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.AddressInfoSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AddressInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Chain":         result.Chain,
		"PublicAddress": result.PublicAddress,
	}).Debug("Address Info")

	return result, nil
}
