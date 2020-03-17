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
	"github.com/shengdoushi/base58"
)

var (
	ErrUserNotFound        = errors.New("User not found")
	ErrInvalidPasswordHash = errors.New("Invalid PasswordHash")
	ErrDatabaseError       = errors.New("Invalid PasswordHash")
)

func saltCredentials(ctx context.Context, login, password string) ([]byte, []byte) {
	loginHash := security.SaltedHash(ctx, []byte(login))
	passwordHash := security.SaltedHash(ctx, []byte(login+password))

	return loginHash, passwordHash
}

func HashEntry(entry string) string {
	hash := utils.HashBytes([]byte(entry))
	return hex.EncodeToString(hash[:])
}

func CreateOrUpdatedCredential(ctx context.Context, database bank.Database, userID uint64, login, password, otpSecret string) (*model.Credential, error) {
	switch db := database.DB().(type) {
	case *gorm.DB:

		// perform a sha512 hex digest of login and password
		login = HashEntry(login)
		password = HashEntry(password)
		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		passwordHash := security.SaltedHash(ctx, []byte(password))
		defer utils.Memzero(loginHash)
		defer utils.Memzero(passwordHash)

		var cred model.Credential
		err := db.
			Where(&model.Credential{UserID: userID}).
			Assign(&model.Credential{
				LoginHash:    base58.Encode(loginHash, base58.BitcoinAlphabet),
				PasswordHash: base58.Encode(passwordHash, base58.BitcoinAlphabet),
				TOTPSecret:   otpSecret,
			}).
			FirstOrCreate(&cred).Error

		return &cred, err

	default:
		return nil, ErrInvalidDatabase
	}
}

func CheckCredential(ctx context.Context, database bank.Database, login, password string) (uint64, bool, error) {
	switch db := database.DB().(type) {
	case *gorm.DB:

		// client should send a sha512 hex digest of the password
		// login = hashEntry(login)
		// password = hashEntry(password)

		password = login + password // password prefixed with login for uniqueness
		loginHash := security.SaltedHash(ctx, []byte(login))
		defer utils.Memzero(loginHash)

		var cred model.Credential
		err := db.
			Where(&model.Credential{LoginHash: base58.Encode(loginHash, base58.BitcoinAlphabet)}).
			First(&cred).Error
		if err != nil {
			return 0, false, ErrDatabaseError
		}
		if cred.UserID == 0 {
			return 0, false, ErrUserNotFound
		}

		passwordHash, err := base58.Decode(cred.PasswordHash, base58.BitcoinAlphabet)
		defer utils.Memzero(passwordHash)
		if err != nil {
			return 0, false, ErrInvalidPasswordHash
		}

		return cred.UserID, security.SaltedHashVerify(ctx, []byte(password), passwordHash), nil

	default:
		return 0, false, ErrInvalidDatabase
	}
}