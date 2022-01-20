package database

import (
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidType                         = errors.New("Invalid Type")
	ErrInvalidStatus                       = errors.New("Invalid Status")
	ErrFiatOperationInfoUpdateNotPermitted = errors.New("OperationInfo Update Not Permitted")
	ErrFiatInvalidOperationInfoID          = errors.New("Invalid OperationInfo")
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

	err := gdb.Create(&operation).Error
	if err != nil {
		return model.FiatOperationInfo{}, err
	}

	return operation, nil
}

func FindFiatOperationInfoByUserIDAndSepa(db bank.Database, userID model.UserID, sepaID model.SepaInfoID) ([]model.FiatOperationInfo, error) {
	list, err := QueryFiatOperationList(db, userID, sepaID, model.FiatOperationStatus("*"))
	if err != nil {
		return []model.FiatOperationInfo{}, err
	}
	return list, nil
}

func FindFiatOperationPendingForUserAndSepa(db bank.Database, userID model.UserID, sepaID model.SepaInfoID) ([]model.FiatOperationInfo, error) {
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
	list, err := FindFiatOperationPendingForUserAndSepa(db, userID, sepaID)
	if err != nil {
		return []model.FiatOperationInfo{}, err
	}

	var result []model.FiatOperationInfo
	for _, operation := range list {
		if operation.Type != "withdrawal" {
			continue
		}
		result = append(result, operation)
	}

	return result, nil
}

func QueryFiatOperationList(db bank.Database, userID model.UserID, sepaID model.SepaInfoID, status model.FiatOperationStatus) ([]model.FiatOperationInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	var filters []func(db *gorm.DB) *gorm.DB
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if sepaID == 0 {
		return nil, ErrInvalidSepaID
	}

	filters = append(filters, ScopeUserID(userID))
	filters = append(filters, ScopeSepaInfoID(sepaID))

	// manage wildcards
	if status != "*" {
		filters = append(filters, ScopeFiatOperationStatus(status))
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

func ScopeFiatOperationStatus(status model.FiatOperationStatus) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqFiatOperationStatus(), status)
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
func reqSepaInfoID() string {
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
