package database

import (
	"errors"

	"git.condensat.tech/bank"

	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidUserID    = errors.New("Invalid UserID")
	ErrInvalidUserName  = errors.New("Invalid User Name")
	ErrInvalidUserEmail = errors.New("Invalid User Email")
)

func FindOrCreateUser(db bank.Database, user model.User) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if len(user.Name) == 0 {
			return model.User{}, ErrInvalidUserName
		}

		if len(user.Email) == 0 {
			return model.User{}, ErrInvalidUserEmail
		}

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

func UserExists(db bank.Database, userID model.UserID) bool {
	entry, err := FindUserById(db, userID)

	return err == nil && entry.ID > 0
}

func UserCount(db bank.Database) (int, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		var result int64
		err := gdb.
			Model(&model.User{}).
			Group("email").
			Count(&result).Error

		return int(result), err

	default:
		return 0, ErrInvalidDatabase
	}
}

func FindUserById(db bank.Database, userID model.UserID) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if userID == 0 {
			return model.User{}, ErrInvalidUserID
		}

		var result model.User
		err := gdb.
			Where(&model.User{ID: userID}).
			First(&result).Error

		return result, err

	default:
		return model.User{}, ErrInvalidDatabase
	}
}

func FindUserByEmail(db bank.Database, email model.UserEmail) (model.User, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if len(email) == 0 {
			return model.User{}, ErrInvalidUserEmail
		}

		var result model.User
		err := gdb.
			Where(&model.User{Email: email}).
			First(&result).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return model.User{}, err
		}

		return result, nil

	default:
		return model.User{}, ErrInvalidDatabase
	}
}
