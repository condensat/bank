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

const (
	withOperatorAuth = true
)

func FiatWithdraw(ctx context.Context, authInfo common.AuthInfo, userName string, withdraw common.AccountEntry, sepaInfo common.FiatOperationInfo) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatWithdraw")

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

	// Look for the user and its accounts
	email := fmt.Sprintf("%s@condensat.tech", userName)

	user, err := database.FindUserByEmail(db, model.UserEmail(email))
	if err != nil {
		return common.AccountEntry{}, err
	}

	if user.ID == 0 {
		return common.AccountEntry{}, errors.New("userID can't be 0")
	}

	// Get AccountID with UserID
	account, err := database.GetAccountsByUserAndCurrencyAndName(db, user.ID, model.CurrencyName(withdraw.Currency), model.AccountName("*"))
	if err != nil || len(account) == 0 {
		return common.AccountEntry{}, errors.New("Account not found")
	}

	withdraw.AccountID = uint64(account[0].ID)

	// Look for the sepa with userID and IBAN
	sepaUser, err := database.GetSepaByUserAndIban(db, user.ID, model.Iban(sepaInfo.IBAN))
	if err != nil && err != database.ErrSepaNotFound {
		return common.AccountEntry{}, err
	}

	if sepaUser.ID == 0 {

		// if sepa is not registered, add it to db
		sepaUser, err = database.CreateSepa(db, model.FiatSepaInfo{
			UserID: user.ID,
			IBAN:   model.Iban(sepaInfo.IBAN),
			BIC:    model.Bic(sepaInfo.BIC),
			Label:  model.String(sepaInfo.Label),
		})
		if err != nil {
			return common.AccountEntry{}, err
		}
	} else {

		// Is there a fiatoperation for this sepa AND this user?
		fiatOperation, err := database.FindFiatWithdrawalPendingForUserAndSepa(db, user.ID, sepaUser.ID)
		if err != nil {
			return common.AccountEntry{}, err
		}

		// stop if there's already 1 or more pending withdrawal
		switch len := len(fiatOperation); len {
		case 0:
			break
		case 1:
			return common.AccountEntry{}, errors.New("Already a pending withdrawal for this user and sepa")
		default:
			return common.AccountEntry{}, errors.New("Multiple pending withdrawals for this user and sepa")
		}
	}

	var withdrawAmount model.Float = model.Float(-withdraw.Amount)
	// If there's no pending withdrawal, let's create the operation
	_, err = database.AddFiatOperationInfo(db, model.FiatOperationInfo{
		UserID:       user.ID,
		SepaInfoID:   sepaUser.ID,
		CurrencyName: model.CurrencyName(withdraw.Currency),
		Amount:       &withdrawAmount,
		Type:         "withdrawal",
		Status:       model.FiatOperationStatusPending,
	})
	if err != nil {
		return common.AccountEntry{}, err
	}

	// Set reference id as userID
	withdraw.ReferenceID = uint64(user.ID) // or SepaID?
	log = log.WithField("ReferenceID", withdraw.ReferenceID)

	// Now do the operation
	result, err := AccountOperation(ctx, withdraw)
	if err != nil {
		return common.AccountEntry{}, errors.New("AccountOperation failed")
	}

	log.WithFields(logrus.Fields{
		"Operation":       result.OperationID,
		"OperationPrevID": result.OperationPrevID,
		"Currency":        withdraw.Currency,
		"Amount":          result.Amount,
		"Balance":         result.Balance,
		"Label":           result.Label,
	}).Debug("FiatWithdraw success")

	return result, err
}

func OnFiatWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnFiatWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.FiatWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			operation, err := FiatWithdraw(ctx, request.AuthInfo, request.UserName, request.Source, request.Destination)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FiatWithdraw")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"AccountID": operation.AccountID,
			})

			log.Info("FiatWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
