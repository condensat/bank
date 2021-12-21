package database

import (
	"errors"

	"git.condensat.tech/bank"

	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidPublicKey = errors.New("Invalid PublicKey")
)

func FindUserPgp(db bank.Database, userID model.UserID) (model.UserPGP, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:
		if userID == 0 {
			return model.UserPGP{}, ErrInvalidUserID
		}

		var result model.UserPGP

		err := gdb.
			Where(model.UserPGP{
				UserID: userID,
			}).
			First(&result).Error

		if err != nil {
			return model.UserPGP{}, err
		}

		return result, err

	default:
		return model.UserPGP{}, ErrInvalidDatabase
	}
}

func AddUserPgp(db bank.Database, userID model.UserID, publicKey model.PgpPublicKey, privateKey model.PgpPrivateKey) (model.UserPGP, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:
		if userID == 0 {
			return model.UserPGP{}, ErrInvalidUserID
		}

		if len(publicKey) == 0 {
			return model.UserPGP{}, ErrInvalidPublicKey
		}

		var result model.UserPGP

		err := gdb.
			Where(model.UserPGP{
				UserID: userID,
			}).
			Assign(model.UserPGP{
				UserID:        userID,
				PublicKey:     publicKey,
				PgpPrivateKey: privateKey,
			}).
			FirstOrCreate(&result).Error

		if err != nil {
			return model.UserPGP{}, err
		}

		return result, err

	default:
		return model.UserPGP{}, ErrInvalidDatabase
	}
}
