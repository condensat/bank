package database

import (
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidWithdrawID     = errors.New("Invalid WithdrawID")
	ErrInvalidWithdrawAmount = errors.New("Invalid Amount")
	ErrInvalidBatchMode      = errors.New("Invalid BatchMode")
)

func AddWithdraw(db bank.Database, from, to model.AccountID, amount model.Float, batch model.BatchMode, data model.WithdrawData) (model.Withdraw, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Withdraw{}, errors.New("Invalid appcontext.Database")
	}

	if from == 0 {
		return model.Withdraw{}, ErrInvalidAccountID
	}
	if to == 0 {
		return model.Withdraw{}, ErrInvalidAccountID
	}
	if from == to {
		return model.Withdraw{}, ErrInvalidAccountID
	}
	if amount <= 0.0 {
		return model.Withdraw{}, ErrInvalidWithdrawAmount
	}
	if len(batch) == 0 {
		return model.Withdraw{}, ErrInvalidBatchMode
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.Withdraw{
		Timestamp: timestamp,
		From:      from,
		To:        to,
		Amount:    &amount,
		Batch:     batch,
		Data:      data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.Withdraw{}, err
	}

	return result, nil
}

func GetWithdraw(db bank.Database, ID model.WithdrawID) (model.Withdraw, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Withdraw{}, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return model.Withdraw{}, ErrInvalidWithdrawID
	}

	var result model.Withdraw
	err := gdb.
		Where(&model.Withdraw{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.Withdraw{}, err
	}

	return result, nil
}
