package database

import (
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidLabel  = errors.New("Invalid Label")
	ErrInvalidIban   = errors.New("Invalid Iban")
	ErrInvalidBic    = errors.New("Invalid Bic")
	ErrInvalidType   = errors.New("Invalid Type")
	ErrInvalidStatus = errors.New("Invalid Status")
)

func AddFiatOperationInfo(db bank.Database, operation model.FiatOperationInfo) (model.FiatOperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.FiatOperationInfo{}, errors.New("Invalid appcontext.Database")
	}

	err := gdb.Create(&operation).Error
	if err != nil {
		return model.FiatOperationInfo{}, err
	}

	return operation, nil
}

func GetFiatOperationInfo(db bank.Database, Label string) (model.FiatOperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.FiatOperationInfo{}, errors.New("Invalid appcontext.Database")
	}

	if len(Label) == 0 {
		return model.FiatOperationInfo{}, ErrInvalidLabel
	}

	var result model.FiatOperationInfo
	err := gdb.
		Where(&model.FiatOperationInfo{Label: Label}).
		First(&result).Error
	if err != nil {
		return model.FiatOperationInfo{}, err
	}

	return result, nil
}
