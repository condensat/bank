package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/security/utils"
	"github.com/sirupsen/logrus"
)

const withOperatorAuth = false

func FiatDeposit(ctx context.Context, authInfo common.AuthInfo, userName string, deposit common.AccountEntry) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatDeposit")

	db := appcontext.Database(ctx)
	if db == nil {
		return common.AccountEntry{}, errors.New("Invalid Database")
	}

	if withOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return common.AccountEntry{}, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return common.AccountEntry{}, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return common.AccountEntry{}, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return common.AccountEntry{}, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return common.AccountEntry{}, errors.New("CheckTOTP failed")
		}
		if !valid {
			return common.AccountEntry{}, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return common.AccountEntry{}, errors.New("Wrong operator ID")
		}
	}

	email := fmt.Sprintf("%s@condensat.tech", userName)

	user, err := database.FindUserByEmail(db, model.UserEmail(email))
	if err != nil {
		return common.AccountEntry{}, err
	}

	if user.ID == 0 {
		return common.AccountEntry{}, errors.New("userID can't be 0")
	}

	// Get AccountID with UserID
	account, err := database.GetAccountsByUserAndCurrencyAndName(db, user.ID, model.CurrencyName(deposit.Currency), model.AccountName("*"))
	if err != nil || len(account) == 0 {
		return common.AccountEntry{}, errors.New("Account not found")
	}

	deposit.AccountID = uint64(account[0].ID)

	log = log.WithField("accountID", deposit.AccountID)

	// Set reference id as userID
	deposit.ReferenceID = uint64(user.ID)
	log = log.WithField("ReferenceID", deposit.ReferenceID)

	// Making the operation
	result, err := AccountOperation(ctx, deposit)
	if err != nil {
		return common.AccountEntry{}, errors.New("AccountOperation failed")
	}

	result.Currency = deposit.Currency

	log.WithFields(logrus.Fields{
		"Operation":       result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Currency":        result.Currency,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
		"Label":           result.Label,
	}).Debug("FiatDeposit success")

	return result, err
}

func OnFiatDeposit(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatDeposit")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatDeposit
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			operation, err := FiatDeposit(ctx, request.AuthInfo, request.UserName, request.Destination)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatDeposit")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"AccountID": operation.AccountID,
			})

			log.Info("FiatDeposit succeeded")

			// create & return response
			return &operation, nil
		})
}
