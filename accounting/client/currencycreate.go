package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencyCreate(ctx context.Context, currencyName string) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyCreate")

	request := common.CurrencyInfo{
		Name: currencyName,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, common.CurrencyCreateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Name":      result.Name,
		"Available": result.Available,
	}).Debug("Currency Created")

	return result, nil
}
