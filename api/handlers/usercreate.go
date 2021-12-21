package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/api/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/security"
	"git.condensat.tech/bank/security/utils"

	"github.com/pquerna/otp/totp"
	"github.com/shengdoushi/base58"
	"github.com/sirupsen/logrus"
)

const (
	withOperatorAuth = true
)

func UserCreate(ctx context.Context, authInfo common.AuthInfo, pgpPublicKey common.PGPPublicKey) (common.UserInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "api.UserCreate")

	db := appcontext.Database(ctx)
	if db == nil {
		return common.UserInfo{}, errors.New("Invalid Database")
	}

	if withOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return common.UserInfo{}, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return common.UserInfo{}, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return common.UserInfo{}, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return common.UserInfo{}, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return common.UserInfo{}, errors.New("CheckTOTP failed")
		}
		if !valid {
			return common.UserInfo{}, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return common.UserInfo{}, errors.New("Wrong operator ID")
		}
	}

	var accountNumber string
	var email string
	for {
		accountNumber = randSeq(common.AccountNumberLength)
		email = fmt.Sprintf("%s@condensat.tech", accountNumber)

		user, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return common.UserInfo{}, errors.New("Database Error")
		}
		// brand new user, break
		if user.ID == 0 {
			break
		}

		// user exists, generate new acount number
		accountNumber = randSeq(common.AccountNumberLength)
	}

	// got an accountNumber, create TOTP credentials
	log = log.WithField("AccountNumber", accountNumber)

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "condensat.tech",
		AccountName: accountNumber,
	})
	if err != nil {
		log.WithError(err).Errorf("Failed to create TOTP")
		return common.UserInfo{}, errors.New("TOTP create Failed")
	}

	var credential model.Credential
	err = db.Transaction(func(tx bank.Database) error {
		// Store user info into database
		user, errTx := database.FindOrCreateUser(tx, model.User{
			Name:  model.UserName(accountNumber),
			Email: model.UserEmail(email),
		})
		if errTx != nil {
			return errTx
		}

		login := database.HashEntry(model.Base58(accountNumber))
		loginHash := security.SaltedHash(ctx, utils.HashString(string(login)))
		passwordHash := security.SaltedHash(ctx, []byte(randSeq(32)))
		credential, errTx = database.CreateOrUpdatedCredential(ctx, tx,
			model.Credential{
				UserID:       user.ID,
				LoginHash:    model.Base58(base58.Encode(loginHash, security.DefaultAlphabet)),
				PasswordHash: model.Base58(passwordHash),                            // password not used for TOTP, genrate random one
				TOTPSecret:   model.String(security.WriteSecret(ctx, key.Secret())), // secret is encrypted with current PasswordHashSeed
			},
		)
		if errTx != nil {
			return errTx
		}

		_, errTx = database.AddUserPgp(tx, user.ID, model.PgpPublicKey(security.WriteSecret(ctx, string(pgpPublicKey))), "")
		if errTx != nil {
			return errTx
		}

		return nil
	})
	if credential.UserID == 0 {
		err = errors.New("Failed to create credential for User")
	}
	if err != nil {
		log.WithError(err).
			Error("Failed to CreateOrUpdatedCredential")
		return common.UserInfo{}, errors.New("Failed to CreateOrUpdatedCredential")
	}

	log.Info("User account created with TOTP credentials")
	return common.UserInfo{
		AccountNumber: accountNumber,
	}, nil
}

func OnUserCreate(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Api.OnUserCreate")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.UserCreation
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {

			user, err := UserCreate(ctx, request.AuthInfo, request.PGPPublicKey)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to create User")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"AccountNumber": user.AccountNumber,
			})

			list, err := client.CurrencyList(ctx)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to list currency")
				return nil, cache.ErrInternalError
			}
			for _, curency := range list.Currencies {
				if !curency.Available {
					continue
				}
				if !curency.AutoCreate {
					continue
				}

				account, err := client.AccountCreate(ctx, user.UserID, curency.Name)
				if err != nil {
					log.WithError(err).
						WithField("Currency", curency.Name).
						Errorf("Failed to create account currency")
					continue
				}
				_, err = client.AccountSetStatus(ctx, account.Info.AccountID, "normal")
				if err != nil {
					log.WithError(err).
						Error("AccountSetStatus Failed")
					continue
				}
				log.
					WithField("Currency", account.Info.Name).
					Debug("User account currency created")
			}

			log.Info("User created with currency account")

			// create & return response
			return &common.UserCreation{
				UserInfo: common.UserInfo{
					// UserID:        user.UserID,
					AccountNumber: user.AccountNumber,
					Timestamp:     user.Timestamp,
					PayLoad:       user.PayLoad,
					// TOTP:          user.TOTP,
				},
			}, nil
		})
}
