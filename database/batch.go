package database

import (
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidBatchID        = errors.New("Invalid BatchID")
	ErrInvalidBatchWithdraws = errors.New("Invalid Withdraws")
)

func AddBatch(db bank.Database, data model.BatchData, withdraws ...model.WithdrawID) (model.Batch, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Batch{}, errors.New("Invalid appcontext.Database")
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.Batch{
		Timestamp: timestamp,
		Data:      data,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.Batch{}, err
	}

	err = AddWithdrawToBatch(db, result.ID, withdraws...)
	if err != nil {
		return model.Batch{}, err
	}

	return result, nil
}

func GetBatch(db bank.Database, ID model.BatchID) (model.Batch, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Batch{}, errors.New("Invalid appcontext.Database")
	}

	if ID == 0 {
		return model.Batch{}, ErrInvalidBatchID
	}

	var result model.Batch
	err := gdb.
		Where(&model.Batch{ID: ID}).
		First(&result).Error
	if err != nil {
		return model.Batch{}, err
	}

	return result, nil
}
