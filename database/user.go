package database

import (
	"context"

	"git.condensat.tech/bank"

	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

func FinddOrCreateUser(ctx context.Context, database bank.Database, name, email string) (*model.User, error) {
	switch db := database.DB().(type) {
	case *gorm.DB:

		result := model.User{
			Name:  name,
			Email: email,
		}
		err := db.
			Where("name = ?", name).
			Where("email = ?", email).
			FirstOrCreate(&result).Error

		return &result, err

	default:
		return nil, ErrInvalidDatabase
	}
}
