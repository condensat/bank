package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencyCreate(ctx context.Context, currencyName, currencyDisplayName string, currencyType common.CurrencyType, isCrypto bool, displayPrecision uint) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.CurrencyCreate")

	request := common.CurrencyInfo{
		Name:             currencyName,
		DisplayName:      currencyDisplayName,
		Available:        false,
		AutoCreate:       false,
		Crypto:           isCrypto,
		Type:             currencyType,
		Asset:            currencyType == 2,
		DisplayPrecision: displayPrecision,
	}

	var result common.CurrencyInfo
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.CurrencyCreateSubject, &request, &result)
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
		"Crypto":           result.Crypto,
		"Type":             result.Type,
		"Asset":            result.Asset,
		"DisplayPrecision": result.DisplayPrecision,
	}).Debug("Currency Created")

	return result, nil
}
