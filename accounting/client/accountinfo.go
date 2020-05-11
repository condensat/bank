package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountInfo(ctx context.Context, accountID uint64) (common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountInfo")

	request := common.AccountInfo{
		AccountID: accountID,
	}

	var result common.AccountInfo
	err := messaging.RequestMessage(ctx, common.AccountInfoSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"AccountID": result.AccountID,
		"Currency":  result.Currency,
		"Name":      result.Name,
		"Status":    result.Status,
	}).Debug("Account Info")

	return result, nil
}
