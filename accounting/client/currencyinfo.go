package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencyInfo(ctx context.Context, currencyName string) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyInfo")

	request := common.CurrencyInfo{
		Name: currencyName,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, common.CurrencyInfoSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.CurrencyInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Name":             result.Name,
		"DisplayName":      result.DisplayName,
		"Available":        result.Available,
		"AutoCreate":       result.AutoCreate,
		"Type":             result.Type,
		"Crypto":           result.Crypto,
		"DisplayPrecision": result.DisplayPrecision,
	}).Trace("Currency CurrencyInfo")

	return result, nil
}
