package database

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

func CreateAccount(ctx context.Context, account model.Account) (model.Account, error) {
	db := appcontext.Database(ctx)
	switch db := db.DB().(type) {
	case *gorm.DB:

		if !UserExists(ctx, account.UserID) {
			return model.Account{}, ErrUserNotFound
		}

		if !CurrencyExists(ctx, account.CurrencyName) {
			return model.Account{}, ErrCurrencyNotFound
		}

		var result model.Account
		err := db.
			Where(model.Account{
				UserID:       account.UserID,
				CurrencyName: account.CurrencyName,
			}).
			Assign(account).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.Account{}, ErrInvalidDatabase
	}
}
