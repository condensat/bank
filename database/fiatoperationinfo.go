package database

import (
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidType                         = errors.New("Invalid Type")
	ErrInvalidStatus                       = errors.New("Invalid Status")
	ErrFiatOperationInfoUpdateNotPermitted = errors.New("OperationInfo Update Not Permitted")
	ErrFiatInvalidOperationInfoID          = errors.New("Invalid OperationInfoID")
	ErrFiatInvalidAccountID                = errors.New("Invalid AccountID")
)

func AddFiatOperationInfo(db bank.Database, operation model.FiatOperationInfo) (model.FiatOperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.FiatOperationInfo{}, errors.New("Invalid appcontext.Database")
	}

	if operation.ID != 0 {
		return model.FiatOperationInfo{}, ErrOperationInfoUpdateNotPermitted
	}

	operation.CreationTimestamp = operation.CreationTimestamp.UTC().Truncate(time.Second)

	err := gdb.Create(&operation).Error
	if err != nil {
		return model.FiatOperationInfo{}, err
	}

	return operation, nil
}

func FiatOperationFinalize(db bank.Database, operationID model.FiatOperationInfoID) (model.FiatOperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.FiatOperationInfo{}, errors.New("Invalid appcontext.Database")
	}

	if operationID == 0 {
		return model.FiatOperationInfo{}, ErrInvalidOperationInfoID
	}

	var operation model.FiatOperationInfo
	err := gdb.
		Where(&model.FiatOperationInfo{ID: operationID}).
		First(&operation).Error
	if err != nil {
		return model.FiatOperationInfo{}, err
	}

	err = gdb.Model(&operation).Update("status", model.FiatOperationStatusComplete).Error
	if err != nil {
		return model.FiatOperationInfo{}, err
	}

	return operation, nil

}

func FindFiatOperationInfoByUserIDAndSepa(db bank.Database, userID model.UserID, sepaID model.SepaInfoID) ([]model.FiatOperationInfo, error) {
	if userID == 0 {
		return []model.FiatOperationInfo{}, ErrInvalidUserID
	}
	if sepaID == 0 {
		return []model.FiatOperationInfo{}, ErrInvalidSepaID
	}

	list, err := QueryFiatOperationList(db, userID, sepaID, model.OperationType("*"), model.FiatOperationStatus("*"))
	if err != nil {
		return []model.FiatOperationInfo{}, err
	}
	return list, nil
}

func FindFiatOperationPendingForUserAndSepa(db bank.Database, userID model.UserID, sepaID model.SepaInfoID) ([]model.FiatOperationInfo, error) {
	if userID == 0 {
		return []model.FiatOperationInfo{}, ErrInvalidUserID
	}
	if sepaID == 0 {
		return []model.FiatOperationInfo{}, ErrInvalidSepaID
	}

	list, err := FindFiatOperationInfoByUserIDAndSepa(db, userID, sepaID)
	if err != nil {
		return []model.FiatOperationInfo{}, err
	}

	var result []model.FiatOperationInfo
	for _, operation := range list {
		if operation.Status == model.FiatOperationStatusPending {
			result = append(result, operation)
		}
	}

	return result, nil
}

func FindFiatWithdrawalPendingForUserAndSepa(db bank.Database, userID model.UserID, sepaID model.SepaInfoID) ([]model.FiatOperationInfo, error) {
	if userID == 0 {
		return []model.FiatOperationInfo{}, ErrInvalidUserID
	}
	if sepaID == 0 {
		return []model.FiatOperationInfo{}, ErrInvalidSepaID
	}

	list, err := FindFiatOperationPendingForUserAndSepa(db, userID, sepaID)
	if err != nil {
		return []model.FiatOperationInfo{}, err
	}

	var result []model.FiatOperationInfo
	for _, operation := range list {
		if operation.Type != model.OperationTypeWithdraw {
			continue
		}
		result = append(result, operation)
	}

	return result, nil
}

func FetchFiatPendingWithdraw(db bank.Database) ([]model.FiatOperationInfo, error) {
	list, err := QueryFiatOperationList(db, model.UserID(0), model.SepaInfoID(0), model.OperationTypeWithdraw, model.FiatOperationStatusPending)
	if err != nil {
		return []model.FiatOperationInfo{}, err
	}

	return list, nil
}

func QueryFiatOperationList(db bank.Database, userID model.UserID, sepaID model.SepaInfoID, operationType model.OperationType, status model.FiatOperationStatus) ([]model.FiatOperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	var filters []func(db *gorm.DB) *gorm.DB

	// manage wildcards
	if userID > 0 {
		filters = append(filters, ScopeUserID(userID))
	}
	if sepaID > 0 {
		filters = append(filters, ScopeFiatOperationSepaInfoID(sepaID))
	}
	if status != "*" {
		filters = append(filters, ScopeFiatOperationStatus(status))
	}
	if operationType != "*" {
		filters = append(filters, ScopeFiatOperationType(operationType))
	}

	var list []*model.FiatOperationInfo
	err := gdb.Model(&model.FiatSepaInfo{}).
		Scopes(filters...).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertFiatOperationList(list), nil
}

// Takes a slice of references to SepaInfo and returns a slice of SepaInfo values
func convertFiatOperationList(list []*model.FiatOperationInfo) []model.FiatOperationInfo {
	var result []model.FiatOperationInfo
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *curr)
	}

	return result[:]
}

func ScopeFiatOperationSepaInfoID(sepaInfoID model.SepaInfoID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqFiatOperationSepaInfoID(), sepaInfoID)
	}
}

func ScopeFiatOperationStatus(status model.FiatOperationStatus) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqFiatOperationStatus(), status)
	}
}

func ScopeFiatOperationType(operationType model.OperationType) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqFiatOperationType(), operationType)
	}
}

const (
	colFiatOperationStatus     = "status"
	colFiatOperationType       = "type"
	colFiatOperationSepaInfoID = "sepa_info_id"
)

func fiatOperationColumnNames() []string {
	return []string{
		colFiatOperationStatus,
		colFiatOperationSepaInfoID,
		colFiatOperationType,
	}
}

// zero allocation requests string for scope
func reqFiatOperationSepaInfoID() string {
	var req [len(colFiatOperationSepaInfoID) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colFiatOperationSepaInfoID)
	copy(req[off:], reqEQ)

	return string(req[:])
}

func reqFiatOperationStatus() string {
	var req [len(colFiatOperationStatus) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colFiatOperationStatus)
	copy(req[off:], reqEQ)

	return string(req[:])
}

func reqFiatOperationType() string {
	var req [len(colFiatOperationType) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colFiatOperationType)
	copy(req[off:], reqEQ)

	return string(req[:])
}
