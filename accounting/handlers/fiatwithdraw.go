package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

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
	withOperatorAuth      = false
	minAmountFiatWithdraw = 20.0
)

func FiatWithdraw(ctx context.Context, authInfo common.AuthInfo, userName string, withdraw common.AccountEntry, sepaInfo common.FiatSepaInfo) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatWithdraw")

	var result common.AccountEntry
	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	if math.Abs(withdraw.Amount) < minAmountFiatWithdraw {
		return result, errors.New("Amount is below the minimum required for a fiat withdraw")
	}

	// check that IBAN is in the correct format
	validIban, err := sepaInfo.IBAN.Valid()
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

	// Look up the currency info
	currency, err := database.GetCurrencyByName(db, model.CurrencyName(withdraw.Currency))
	if err != nil {
		return result, err
	}

	// Get AccountID with UserID
	account, err := database.GetAccountsByUserAndCurrencyAndName(db, user.ID, model.CurrencyName(currency.Name), model.AccountName("*"))
	if err != nil || len(account) == 0 {
		return result, errors.New("Account not found")
	}

	withdraw.AccountID = uint64(account[0].ID)

	// Look for the sepa with userID and IBAN
	sepaUser, err := database.GetSepaByUserAndIban(db, user.ID, model.Iban(sepaInfo.IBAN))
	if err != nil && err != database.ErrSepaNotFound {
		return result, err
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
			return result, err
		}
	} else {

		// Is there a fiatoperation for this sepa AND this user?
		fiatOperation, err := database.FindFiatWithdrawalPendingForUserAndSepa(db, user.ID, sepaUser.ID)
		if err != nil {
			return result, err
		}

		// stop if there's already 1 or more pending withdrawal
		switch len := len(fiatOperation); len {
		case 0:
			break
		case 1:
			return result, errors.New("Already a pending withdrawal for this user and sepa")
		default:
			return result, errors.New("Multiple pending withdrawals for this user and sepa")
		}
	}

	// Set reference id as userID
	withdraw.ReferenceID = uint64(user.ID) // or SepaID?
	log = log.WithField("ReferenceID", withdraw.ReferenceID)

	// Database Query
	err = db.Transaction(func(db bank.Database) error {
		// How much to pay in fees?
		feeInfo, err := database.GetFeeInfo(db, currency.Name)
		if err != nil {
			log.WithError(err).
				Error("GetFeeInfo failed")
			return err
		}
		if !feeInfo.IsValid() {
			log.Error("Invalid FeeInfo")
			return err
		}

		feeAmount := feeInfo.Compute(model.Float(withdraw.Amount))

		if feeAmount < 0 {
			feeAmount = -feeAmount
		}

		if feeAmount < 0.01 {
			feeAmount = 0.01
		}

		log = log.WithField("feeAmount", feeAmount)

		feeBankAccountID, err := getBankWithdrawAccount(ctx, withdraw.Currency)
		if err != nil {
			return errors.New("Can't get bank account id")
		}

		timestamp := time.Now()
		feeTransfer := common.AccountTransfer{
			Source: common.AccountEntry{
				AccountID: withdraw.AccountID,

				OperationType:    string(model.OperationTypeTransferFee),
				SynchroneousType: "sync",
				ReferenceID:      withdraw.ReferenceID,

				Timestamp: timestamp,

				Amount: float64(-feeAmount),

				Currency: withdraw.Currency,
			},
			Destination: common.AccountEntry{
				AccountID: uint64(feeBankAccountID),

				OperationType:    string(model.OperationTypeTransferFee),
				SynchroneousType: "sync",
				ReferenceID:      withdraw.ReferenceID,

				Timestamp: timestamp,

				Amount: float64(feeAmount),

				Currency: withdraw.Currency,
			},
		}

		_, err = AccountTransferWithDatabase(ctx, db, feeTransfer)
		if err != nil {
			return errors.New("AccountOperation failed")
		}

		// Since we're withdrawing, put a minus sign before amount
		withdraw.Amount = -withdraw.Amount
		log.Debugf("Amount sent to AccountOperation: %v\n", withdraw.Amount)
		// Now do the operation
		result, err = AccountOperation(ctx, withdraw)
		if err != nil {
			return errors.New("AccountOperation failed")
		}

		// switch amount back to positive
		result.Amount = -result.Amount // God that's ugly

		var withdrawAmount model.Float = model.Float(result.Amount)
		_, err = database.AddFiatOperationInfo(db, model.FiatOperationInfo{
			UserID:       user.ID,
			SepaInfoID:   sepaUser.ID,
			CurrencyName: model.CurrencyName(withdraw.Currency),
			Amount:       &withdrawAmount,
			Type:         model.OperationTypeWithdraw,
			Status:       model.FiatOperationStatusPending,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return result, err
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
