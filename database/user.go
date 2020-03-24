package database

import (
	"git.condensat.tech/bank"

	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

func FindOrCreateUser(db bank.Database, user model.User) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result model.User
		err := gdb.
			Where(model.User{
				Name:  user.Name,
				Email: user.Email,
			}).
			Assign(user).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.User{}, ErrInvalidDatabase
	}
}

func UserExists(db bank.Database, userID uint64) bool {
	entry, err := FindUserById(db, userID)

	return err == nil && entry.ID > 0
}

func FindUserById(db bank.Database, userID uint64) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result model.User
		err := gdb.
			Where(&model.User{ID: userID}).
			First(&result).Error

		return result, err

	default:
		return model.User{}, ErrInvalidDatabase
	}
}
