package database

import (
	"context"
	"encoding/hex"
	"errors"

	"git.condensat.tech/bank"

	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/security"
	"git.condensat.tech/bank/security/utils"

	"github.com/jinzhu/gorm"
	"github.com/pquerna/otp/totp"
	"github.com/shengdoushi/base58"
)

var (
	ErrUserNotFound        = errors.New("User not found")
	ErrInvalidPasswordHash = errors.New("Invalid PasswordHash")
	ErrDatabaseError       = errors.New("Database Operation failed")
	ErrInvalidTOTP         = errors.New("Invalid TOTP")
)

func HashEntry(entry model.Base58) model.Base58 {
	hash := utils.HashBytes([]byte(entry))
	return model.Base58(hex.EncodeToString(hash[:]))
}

func CreateOrUpdatedCredential(ctx context.Context, db bank.Database, credential model.Credential) (model.Credential, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		// perform a sha512 hex digest of login and password
		login := HashEntry(credential.LoginHash)
		password := HashEntry(credential.PasswordHash)
		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		passwordHash := security.SaltedHash(ctx, []byte(password))
		defer utils.Memzero(loginHash)
		defer utils.Memzero(passwordHash)

		var result model.Credential
		err := gdb.
			Where(&model.Credential{UserID: credential.UserID}).
			Assign(&model.Credential{
				LoginHash:    model.Base58(credential.LoginHash),
				PasswordHash: model.Base58(base58.Encode(passwordHash, base58.BitcoinAlphabet)),
				TOTPSecret:   credential.TOTPSecret,
			}).
			FirstOrCreate(&result).Error

		return result, err

	default:
		return model.Credential{}, ErrInvalidDatabase
	}
}

func CheckCredential(ctx context.Context, db bank.Database, login, password model.Base58) (model.UserID, bool, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		// client should send a sha512 hex digest of the password
		// login = hashEntry(login)
		// password = hashEntry(password)

		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		defer utils.Memzero(loginHash)

		var cred model.Credential
		err := gdb.
			Where(&model.Credential{LoginHash: model.Base58(base58.Encode(loginHash, base58.BitcoinAlphabet))}).
			First(&cred).Error
		if err != nil {
			return 0, false, ErrDatabaseError
		}
		if cred.UserID == 0 {
			return 0, false, ErrUserNotFound
		}

		passwordHash, err := base58.Decode(string(cred.PasswordHash), base58.BitcoinAlphabet)
		defer utils.Memzero(passwordHash)
		if err != nil {
			return 0, false, ErrInvalidPasswordHash
		}

		return cred.UserID, security.SaltedHashVerify(ctx, []byte(password), passwordHash), nil

	default:
		return 0, false, ErrInvalidDatabase
	}
}

func CheckTOTP(ctx context.Context, db bank.Database, login model.Base58, passcode string) (model.UserID, bool, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		loginHash := security.SaltedHash(ctx, utils.HashString(string(login)))

		var cred model.Credential
		err := gdb.
			Where(&model.Credential{
				LoginHash: model.Base58(base58.Encode(loginHash, base58.BitcoinAlphabet)),
			}).
			First(&cred).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return 0, false, ErrDatabaseError
		}

		if len(cred.TOTPSecret) == 0 {
			return 0, false, errors.New("Invalid TOTP")
		}

		secret := security.ReadSecret(ctx, string(cred.TOTPSecret))
		valid := totp.Validate(passcode, secret)

		return cred.UserID, valid, nil

	default:
		return 0, false, ErrInvalidDatabase
	}
}
