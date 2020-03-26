package database

import (
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidAccountOperation = errors.New("Invalid Account Operation")
)

func AppendAccountOperation(db bank.Database, operation model.AccountOperation) (model.AccountOperation, error) {
	var result model.AccountOperation
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	// check for valid accountID
	accountID := operation.AccountID
	if accountID == 0 {
		return result, ErrInvalidAccountID
	}

	// UTC timestamp
	operation.Timestamp = operation.Timestamp.UTC().Truncate(time.Second)

	// pre-check operation with ids
	if !operation.PreCheck() {
		return result, ErrInvalidAccountOperation
	}

	// within a db transaction
	// returning error will cause rollback
	err := db.Transaction(func(db bank.Database) error {

		// get Account (for currency)
		account, err := GetAccountByID(db, accountID)
		if err != nil {
			return ErrAccountNotFound
		}

		// check currency status
		curr, err := GetCurrencyByName(db, account.CurrencyName)
		if err != nil {
			return ErrCurrencyNotFound
		}
		if !curr.IsAvailable() {
			return ErrCurrencyNotAvailable
		}

		// check account status
		accountState, err := GetAccountStatusByAccountID(db, accountID)
		if err != nil {
			return ErrAccountStateNotFound
		}
		if !accountState.State.Valid() {
			return ErrInvalidAccountState
		}

		// update PrevID with last operation ID
		previousOperation, err := GetLastAccountOperation(db, accountID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		operation.PrevID = previousOperation.ID

		// store operation
		gdb := getGormDB(db)
		if gdb != nil {
			err = gdb.Create(&operation).Error
			if err != nil {
				return err
			}
			// check if operation is valid
			if !operation.IsValid() {
				return ErrInvalidAccountOperation
			}

			// get result and commit transaction
			result = operation
		}

		return nil
	})

	// return result with error
	return result, err
}

func GetLastAccountOperation(db bank.Database, accountID model.AccountID) (model.AccountOperation, error) {
	var result model.AccountOperation

	gdb := getGormDB(db)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if accountID == 0 {
		return result, ErrInvalidAccountID
	}

	err := gdb.
		Where(model.AccountOperation{
			AccountID: accountID,
		}).
		Last(&result).Error

	return result, err
}
