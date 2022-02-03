package handlers

import (
	"context"
	"errors"
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
	"github.com/sirupsen/logrus"
)

const (
	minAmountFiatWithdraw = 20.0
)

func FiatWithdraw(ctx context.Context, userId uint64, withdraw common.AccountEntry, sepaInfo common.FiatSepaInfo) (common.AccountEntry, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.FiatWithdraw")
	var result common.AccountEntry

	// Sanity checks
	if userId == 0 {
		return result, errors.New("userID can't be 0")
	}

	if withdraw.Amount < minAmountFiatWithdraw {
		log.WithField("Amount", withdraw.Amount).
			Error("Insufficient amount")
		return result, errors.New("Amount is below the minimum required for a fiat withdraw")
	}

	// check that IBAN is in the correct format
	validIban, err := sepaInfo.IBAN.Valid()
	if err != nil {
		return result, err
	}
	if !validIban {
		log.WithField("IBAN", sepaInfo.IBAN).Error("Provided iban doesn't respect format")
		return result, cache.ErrInternalError
	}

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// Look up the currency info
	currency, err := database.GetCurrencyByName(db, model.CurrencyName(withdraw.Currency))
	if err != nil {
		return result, err
	}

	// Round up to currency.DisplayPrecision
	rounding := math.Pow10(int(*currency.Precision))
	withdraw.Amount = math.Floor(withdraw.Amount*rounding) / rounding

	// Compute fees amount and find bank account to pay to
	feeInfo, err := database.GetFeeInfo(db, currency.Name)
	if err != nil {
		log.WithError(err).
			Error("GetFeeInfo failed")
		return result, err
	}
	if !feeInfo.IsValid() {
		log.Error("Invalid FeeInfo")
		return result, err
	}

	feeAmount := feeInfo.Compute(model.Float(withdraw.Amount))

	if feeAmount < 0 {
		return result, errors.New("Negative fee amount are not allowed")
	}

	if feeAmount < feeInfo.Minimum {
		feeAmount = feeInfo.Minimum
	}

	log = log.WithField("feeAmount", feeAmount)

	feeBankAccountID, err := getBankWithdrawAccount(ctx, withdraw.Currency)
	if err != nil {
		return result, errors.New("Can't get bank account id")
	}

	// Get AccountID with UserID only if no accountId provided
	if withdraw.AccountID == 0 {
		accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, model.UserID(userId), model.CurrencyName(currency.Name), model.AccountName("*"))
		if err != nil || len(accounts) == 0 {
			return result, errors.New("Accounts not found")
		}

		for _, account := range accounts {
			// get account info
			accountInfo, err := AccountInfo(ctx, uint64(account.ID))
			if err != nil {
				return result, err
			}
			if accountInfo.Status != "normal" {
				continue
			}

			// Check available balance too
			if (withdraw.Amount + float64(feeAmount)) > accountInfo.Balance {
				continue
			}

			// We found a suitable account
			withdraw.AccountID = uint64(account.ID)
			break
		}
	}

	if withdraw.AccountID == 0 {
		return result, errors.New("Can't find an account that allows withdraw for this user and currency")
	}

	// Look for the sepa with userID and IBAN
	sepaUser, err := database.GetSepaByUserAndIban(db, model.UserID(userId), model.Iban(sepaInfo.IBAN))
	if err != nil && err != database.ErrSepaNotFound {
		return result, err
	}

	if sepaUser.ID == 0 {

		// if sepa is not registered, add it to db
		sepaUser, err = database.CreateSepa(db, model.FiatSepaInfo{
			UserID: model.UserID(userId),
			IBAN:   model.Iban(sepaInfo.IBAN),
			BIC:    model.Bic(sepaInfo.BIC),
			Label:  model.String(sepaInfo.Label),
		})
		if err != nil {
			return result, err
		}

	} else {

		// Is there a fiatoperation for this sepa AND this user?
		fiatOperation, err := database.FindFiatWithdrawalPendingForUserAndSepa(db, model.UserID(userId), sepaUser.ID)
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

	// Set referenceId as sepaId
	withdraw.ReferenceID = uint64(sepaUser.ID)
	log = log.WithField("ReferenceID", withdraw.ReferenceID)

	// Database Query
	err = db.Transaction(func(db bank.Database) error {

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
		// Now do the operation
		result, err = AccountOperation(ctx, withdraw)
		if err != nil {
			return errors.New("AccountOperation failed")
		}

		// switch amount back to positive
		result.Amount = -result.Amount // God that's ugly

		// Add the currency to the result
		result.Currency = withdraw.Currency

		withdrawAmount := model.Float(result.Amount)
		_, err = database.AddFiatOperationInfo(db, model.FiatOperationInfo{
			UserID:       model.UserID(userId),
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
			operation, err := FiatWithdraw(ctx, request.UserId, request.Source, request.Destination)
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
