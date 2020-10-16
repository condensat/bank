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

func AccountSetStatus(ctx context.Context, accountID uint64, state string) (common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountSetStatus")

	request := common.AccountInfo{
		AccountID: accountID,
		Status:    state,
	}

	var result common.AccountInfo
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.AccountSetStatusSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"AccountID": result.AccountID,
		"Status":    result.Status,
	}).Debug("Account SetStatus")

	return result, nil
}
