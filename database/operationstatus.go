package database

import (
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidOperationStatus = errors.New("Invalid OperationStatus")
)

// AddOrUpdateOperationStatus
func AddOrUpdateOperationStatus(db bank.Database, operation model.OperationStatus) (model.OperationStatus, error) {
	var result model.OperationStatus
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	if operation.OperationInfoID == 0 {
		return result, ErrInvalidOperationInfoID
	}

	if len(operation.State) == 0 {
		return result, ErrInvalidOperationStatus
	}

	operation.LastUpdate = time.Now().UTC().Truncate(time.Second)

	err := gdb.
		Where(model.OperationStatus{
			OperationInfoID: operation.OperationInfoID,
		}).
		Assign(operation).
		FirstOrCreate(&result).Error

	return result, err
}

type DepositInfos struct {
	Count  int
	Active int
}

func DepositsInfos(db bank.Database) (DepositInfos, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return DepositInfos{}, errors.New("Invalid appcontext.Database")
	}

	var totalOperations int64
	err := gdb.Model(&model.OperationStatus{}).
		Count(&totalOperations).Error
	if err != nil {
		return DepositInfos{}, err
	}

	var activeOperations int64
	err = gdb.Model(&model.OperationStatus{}).
		Where("state <> ?", "settled").
		Count(&activeOperations).Error
	if err != nil {
		return DepositInfos{}, err
	}

	return DepositInfos{
		Count:  int(totalOperations),
		Active: int(activeOperations),
	}, nil
}

// GetOperationStatus
func GetOperationStatus(db bank.Database, operationInfoID model.OperationInfoID) (model.OperationStatus, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.OperationStatus{}, errors.New("Invalid appcontext.Database")
	}

	if operationInfoID == 0 {
		return model.OperationStatus{}, ErrInvalidOperationInfoID
	}

	var result model.OperationStatus
	err := gdb.
		Where(model.OperationStatus{
			OperationInfoID: operationInfoID,
		}).
		First(&result).Error
	if err != nil {
		return model.OperationStatus{}, err
	}

	return result, nil
}

func FindActiveOperationStatus(db bank.Database) ([]model.OperationStatus, error) {
	gdb := db.DB().(*gorm.DB)

	var list []*model.OperationStatus
	err := gdb.
		Where("state <> ?", "settled").
		Or("accounted <> ?", "settled").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertOperationStatusList(list), nil
}

func FindActiveOperationInfo(db bank.Database) ([]model.OperationInfo, error) {
	gdb := db.DB().(*gorm.DB)

	subQueryState := gdb.Model(&model.OperationStatus{}).
		Select("operation_info_id").
		Where("state <> ?", "settled").
		Or("accounted <> ?", "settled").
		SubQuery()

	var list []*model.OperationInfo
	err := gdb.Model(&model.OperationInfo{}).
		Joins("JOIN (?) AS os ON operation_info.id = os.operation_info_id", subQueryState).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertOperationInfoList(list), nil
}

func convertOperationStatusList(list []*model.OperationStatus) []model.OperationStatus {
	var result []model.OperationStatus
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *curr)
	}

	return result[:]
}
