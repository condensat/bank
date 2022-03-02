package database

import (
	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

func AddOperator(db bank.Database, operator model.Operator) (model.Operator, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return model.Operator{}, ErrInvalidDatabase
	}

	// store operation
	err := gdb.Create(&operator).Error
	if err != nil {
		return model.Operator{}, err
	}

	return operator, nil
}

func GetOperatorById(db bank.Database, ID model.OperatorID) (model.Operator, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return model.Operator{}, ErrInvalidDatabase
	}

	var result model.Operator

	err := gdb.Model(&result).
		Scopes(ScopeOperatorID(ID)).
		First(&result).Error
	if err != nil {
		return result, err
	}

	return result, nil
}

func UpdateOperator(db bank.Database, update model.Operator) (model.Operator, error) {
	// Get the Operator entry
	operator, err := GetOperatorById(db, update.ID)
	if err != nil {
		return model.Operator{}, err
	}

	err = db.Transaction(func(tx bank.Database) error {
		if update.AccountOperationID != 0 {
			operator, err = updateOperatorAccountOperation(db, operator, update.AccountOperationID)
			if err != nil {
				return err
			}
		}

		if update.AccountID != 0 {
			operator, err = updateOperatorAccount(db, operator, update.AccountID)
			if err != nil {
				return err
			}
		}

		return err
	})

	return operator, nil
}

func updateOperatorAccount(db bank.Database, operator model.Operator, accountID model.AccountID) (model.Operator, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return model.Operator{}, ErrInvalidDatabase
	}

	// Update AccountId column
	err := gdb.Model(&operator).Update("account_id", accountID).Error
	if err != nil {
		return operator, err
	}

	return operator, nil
}

func updateOperatorAccountOperation(db bank.Database, operator model.Operator, accountOperationID model.AccountOperationID) (model.Operator, error) {
	gdb := getGormDB(db)
	if gdb == nil {
		return model.Operator{}, ErrInvalidDatabase
	}

	// Update AccountOperationId column
	err := gdb.Model(&operator).Update("account_operation_id", accountOperationID).Error
	if err != nil {
		return operator, err
	}

	return operator, nil
}

func ScopeOperatorID(operatorID model.OperatorID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqOperatorID(), operatorID)
	}
}

const (
	colOperatorID = "id"
)

func operatorColumnNames() []string {
	return []string{
		colOperatorID,
	}
}

func reqOperatorID() string {
	var req [len(colOperatorID) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colOperatorID)
	copy(req[off:], reqEQ)

	return string(req[:])
}
