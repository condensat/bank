package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountCreate(ctx context.Context, userID uint64, currency string) (common.AccountCreation, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountCreate")

	request := common.AccountCreation{
		UserID: userID,
		Info: common.AccountInfo{
			Currency: common.CurrencyInfo{
				Name: currency,
			},
		},
	}

	var result common.AccountCreation
	err := messaging.RequestMessage(ctx, common.AccountCreateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountCreation{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"UserID":    result.UserID,
		"AccountID": result.Info.AccountID,
	}).Debug("Account Created")

	return result, nil
}
