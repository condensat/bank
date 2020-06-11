package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CurrencySetAvailable(ctx context.Context, currencyName string, available bool) (common.CurrencyInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CurrencySetAvailable")
	var result common.CurrencyInfo

	log = log.WithField("CurrencyName", currencyName)

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		// check if currency exists
		currency, err := database.GetCurrencyByName(db, model.CurrencyName(currencyName))
		if err != nil {
			log.WithError(err).Error("Failed to GetCurrencyByName")
			return err
		}

		if string(currency.Name) != currencyName {
			return database.ErrCurrencyNotFound
		}

		if currency.IsAvailable() == available {
			// NOOP
			result = common.CurrencyInfo{
				Name:             string(currency.Name),
				DisplayName:      string(currency.DisplayName),
				Available:        currency.IsAvailable(),
				AutoCreate:       currency.AutoCreate,
				Crypto:           currency.IsCrypto(),
				Type:             common.CurrencyType(currency.GetType()),
				Asset:            currency.GetType() == 2,
				DisplayPrecision: uint(currency.DisplayPrecision()),
			}
			return nil
		}

		var availableState int
		if available {
			availableState = 1
		}

		var crypto model.Int
		if currency.IsCrypto() {
			crypto = 1
		}

		// update currency available
		currency, err = database.AddOrUpdateCurrency(db,
			model.NewCurrency(
				model.CurrencyName(currencyName),
				model.CurrencyName(currency.DisplayName),
				model.Int(currency.GetType()),
				model.Int(availableState),
				crypto,
				currency.DisplayPrecision(),
			),
		)
		if err != nil {
			log.WithError(err).Error("Failed to AddOrUpdateCurrency")
			return err
		}

		result = common.CurrencyInfo{
			Name:             string(currency.Name),
			DisplayName:      string(currency.DisplayName),
			Available:        currency.IsAvailable(),
			AutoCreate:       currency.AutoCreate,
			Crypto:           currency.IsCrypto(),
			Type:             common.CurrencyType(currency.GetType()),
			Asset:            currency.GetType() == 2,
			DisplayPrecision: uint(currency.DisplayPrecision()),
		}

		return nil
	})

	if err == nil {
		log.WithFields(logrus.Fields{
			"Name":        result.Name,
			"DisplayName": result.DisplayName,
			"Available":   result.Available,
			"AutoCreate":  result.AutoCreate,
			"Type":        result.Type,
			"Asset":       result.Asset,
			"Crypto":      result.Crypto,
		}).Warn("Currency updated")
	}

	return result, err
}

func OnCurrencySetAvailable(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Currencying.OnCurrencySetAvailable")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CurrencyInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Name": request.Name,
			})

			currency, err := CurrencySetAvailable(ctx, request.Name, request.Available)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CurrencySetAvailable")
				return nil, cache.ErrInternalError
			}

			log.Info("Currency updated")
			// create & return response
			result := currency
			return &result, nil
		})
}