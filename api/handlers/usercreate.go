package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/api/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

const (
	withOperatorAuth = true
)

func UserCreate(ctx context.Context, authInfo common.AuthInfo, pgpPublicKey common.PGPPublicKey) (common.UserInfo, error) {
	return common.UserInfo{}, nil
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
