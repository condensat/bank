package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencySetAvailable(ctx context.Context, currencyName string, available bool) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencySetAvailable")

	request := common.CurrencyInfo{
		Name:      currencyName,
		Available: available,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CurrencySetAvailableSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Name":      result.Name,
		"Available": result.Available,
	}).Debug("Currency SetAvailable")

	return result, nil
}
