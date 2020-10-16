// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountList(ctx context.Context, userID uint64) (common.UserAccounts, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountList")

	request := common.UserAccounts{
		UserID: userID,
	}

	var result common.UserAccounts
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.AccountListSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.UserAccounts{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"UserID": result.UserID,
		"Count":  len(result.Accounts),
	}).Debug("User Accounts")

	return result, nil
}
