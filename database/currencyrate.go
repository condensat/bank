package database

import (
	"context"
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"

	"github.com/jinzhu/gorm"
)

func AppendCurencyRates(ctx context.Context, currencyRates []model.CurrencyRate) error {
	log := logger.Logger(ctx).WithField("Method", "database.AppendCurencyRates")
	db := appcontext.Database(ctx)
	if db == nil {
		return errors.New("Invalid appcontext.Database")
	}

	return db.Transaction(func(tx bank.Database) error {
		txdb := tx.DB().(*gorm.DB)
		if txdb == nil {
			return errors.New("Invalid tx Database")
		}

		var resultErr error
		for _, rate := range currencyRates {
			err := txdb.Create(&rate).Error
			if err != nil {
				log.WithError(err).Warning("Failed to add CurrencyRate")
				resultErr = err // return only last error
				continue        // continue to insert if possible
			}
		}

		return resultErr
	})
}
