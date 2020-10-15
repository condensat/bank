package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencyList(ctx context.Context) (common.CurrencyList, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyList")

	var result common.CurrencyList
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CurrencyListSubject, &result, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyList{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(result.Currencies),
	}).Debug("Currency Created")

	return result, nil
}
