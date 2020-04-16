package client

import (
	"context"
	"time"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountHistory(ctx context.Context, accountID uint64, from, to time.Time) (common.AccountHistory, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.AccountHistory")

	request := common.AccountHistory{
		AccountID: accountID,
		From:      from,
		To:        to,
	}

	var result common.AccountHistory
	err := messaging.RequestMessage(ctx, common.AccountHistorySubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountHistory{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"AccountID": result.AccountID,
		"Count":     len(result.Entries),
	}).Debug("Account History")

	return result, nil
}
