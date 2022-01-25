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

func FiatFinalizeWithdraw(ctx context.Context, authInfo common.AuthInfo, userName string, iban common.IBAN) (common.FiatFinalizeWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatFinalizeWithdraw")

	var result common.FiatFinalizeWithdraw
	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// check that IBAN is in the correct format
	validIban, err := iban.Valid()
	if err != nil {
		return result, err
	}
	if !validIban {
		return result, errors.New("Provided iban doesn't respect format")
	}

	if withOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return result, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return result, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return result, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return result, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return result, errors.New("CheckTOTP failed")
		}
		if !valid {
			return result, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return result, errors.New("Wrong operator ID")
		}
	}

	// Look for the user and its accounts
	email := fmt.Sprintf("%s@condensat.tech", userName)

	user, err := database.FindUserByEmail(db, model.UserEmail(email))
	if err != nil {
		return result, err
	}

	if user.ID == 0 {
		return result, errors.New("userID can't be 0")
	}

	// Look for the sepa with userID and IBAN
	sepaUser, err := database.GetSepaByUserAndIban(db, user.ID, model.Iban(iban))
	if err != nil {
		return result, err
	}

	// Is there a fiatoperation for this sepa AND this user?
	fiatOperation, err := database.FindFiatWithdrawalPendingForUserAndSepa(db, user.ID, sepaUser.ID)
	if err != nil {
		return result, err
	}

	// stop if there's not exactly one pending operation
	switch len := len(fiatOperation); len {
	case 0:
		return result, errors.New("There's no pending withdrawal for this user and sepa")
	case 1:
		break
	default:
		return result, errors.New("Multiple pending withdrawals for this user and sepa")
	}

	// Now we only need to update the status of the fiat Operation
	toUpdate := fiatOperation[0].ID

	var updated model.FiatOperationInfo
	updated, err = database.FiatOperationFinalize(db, toUpdate)
	if err != nil {
		return result, err
	}

	result.UserName = userName
	result.IBAN = iban
	result.Currency = string(updated.CurrencyName)
	result.Amount = float64(*(updated.Amount))

	log.WithFields(logrus.Fields{
		"Currency": result.Currency,
		"Amount":   result.Amount,
		"UserName": result.UserName,
	}).Debug("FiatFinalizeWithdraw success")

	return result, err
}

func OnFiatFinalizeWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatFinalizeWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatFinalizeWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			operation, err := FiatFinalizeWithdraw(ctx, request.AuthInfo, request.UserName, request.IBAN)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatFinalizeWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("FiatFinalizeWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
