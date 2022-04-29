package database

import (
	"errors"
	"math"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

func AddWithdrawValidation(db bank.Database, userID model.UserID, withdrawID model.WithdrawID, base model.CurrencyName, amount model.Float) (model.Validation, error) {
	// Sanity check
	if userID == 0 {
		return model.Validation{}, ErrInvalidUserID
	}

	if withdrawID == 0 {
		return model.Validation{}, ErrInvalidWithdrawID
	}

	if len(base) == 0 {
		return model.Validation{}, ErrInvalidCurrencyName
	}

	if amount <= 0 {
		return model.Validation{}, ErrInvalidOperationAmount
	}

	// Round amount to the precision number for the currency
	precision, err := GetCurrencyPrecision(db, base)
	if err != nil {
		return model.Validation{}, err
	}

	factor := math.Pow(10, float64(precision))
	roundedAmt := math.Floor(float64(amount)*factor) / factor

	// Get the withdraw target type
	target, err := GetWithdrawTargetByWithdrawID(db, withdrawID)
	if err != nil {
		return model.Validation{}, err
	}

	// Write a Validation entry to db
	result, err := newValidation(db, userID, model.OperationTypeWithdraw, model.RefID(withdrawID), target.Type, base, model.Float(roundedAmt))
	if err != nil {
		return model.Validation{}, err
	}

	return result, nil
}

func newValidation(db bank.Database, userID model.UserID, operationType model.OperationType, referenceID model.RefID, targetType model.WithdrawTargetType, base model.CurrencyName, amount model.Float) (model.Validation, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Validation{}, errors.New("Invalid appcontext.Database")
	}

	timestamp := time.Now().UTC().Truncate(time.Second)
	result := model.Validation{
		UserID:        userID,
		Timestamp:     timestamp,
		OperationType: operationType,
		ReferenceID:   referenceID,
		Type:          targetType,
		Base:          base,
		Amount:        &amount,
	}

	err := gdb.Create(&result).Error
	if err != nil {
		return model.Validation{}, err
	}

	return result, nil
}

func GetWithdrawValidationsFromStartToNow(db bank.Database, userID model.UserID, start time.Time, target model.WithdrawTargetType) ([]model.Validation, error) {
	var result []model.Validation

	if userID == 0 {
		return result, ErrInvalidUserID
	}

	if start.After(time.Now()) {
		return result, errors.New("start time can't be in the future")
	}

	end := time.Now().Truncate(time.Second)

	result, err := getValidationsFromRange(db, userID, start, end, model.OperationTypeWithdraw, target)
	if err != nil {
		return result, err
	}

	return result, nil
}

func getValidationsFromRange(db bank.Database, userID model.UserID, start time.Time, end time.Time, operation model.OperationType, target model.WithdrawTargetType) ([]model.Validation, error) {
	var result []model.Validation

	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if start.After(end) {
		start, end = end, start
	}

	var list []*model.Validation
	err := gdb.
		Where(model.Validation{
			UserID:        userID,
			Type:          target,
			OperationType: operation,
		}).
		Where("timestamp BETWEEN ? AND ?", start, end).
		Order("id ASC").
		Limit(HistoryMaxOperationCount).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertValidationList(list), nil
}

func convertValidationList(list []*model.Validation) []model.Validation {
	var result []model.Validation
	for _, curr := range list {
		if curr != nil {
			result = append(result, *curr)
		}
	}

	return result[:]
}
