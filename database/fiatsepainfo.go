package database

import (
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var (
	ErrSepaNoIban    = errors.New("Must provide an Iban")
	ErrSepaNoBic     = errors.New("Must provide a Bic")
	ErrSepaExists    = errors.New("Sepa info Exists")
	ErrSepaNotFound  = errors.New("Sepa info Not Found")
	ErrInvalidSepaID = errors.New("Invalid sepa ID")
)

func CreateSepa(db bank.Database, sepaInfo model.FiatSepaInfo) (model.FiatSepaInfo, error) {
	switch gdb := db.DB().(type) {
	case *gorm.DB:

		if sepaInfo.UserID == 0 {
			return model.FiatSepaInfo{}, ErrInvalidUserID
		}

		if len(sepaInfo.IBAN) == 0 {
			return model.FiatSepaInfo{}, ErrSepaNoIban
		}

		if len(sepaInfo.BIC) == 0 {
			return model.FiatSepaInfo{}, ErrSepaNoBic
		}

		if SepaExists(db, sepaInfo.UserID, sepaInfo.IBAN) {
			return model.FiatSepaInfo{}, ErrSepaExists
		}

		var result model.FiatSepaInfo
		err := gdb.
			Where(model.FiatSepaInfo{
				UserID: sepaInfo.UserID,
				IBAN:   sepaInfo.IBAN,
				BIC:    sepaInfo.BIC,
				Label:  sepaInfo.Label,
			}).
			Assign(sepaInfo).
			FirstOrCreate(&result).Error

		if err != nil {
			return model.FiatSepaInfo{}, err
		}

		return result, err

	default:
		return model.FiatSepaInfo{}, ErrInvalidDatabase
	}
}

func SepaExists(db bank.Database, userID model.UserID, iban model.Iban) bool {
	_, err := GetSepaByUserAndIban(db, userID, iban)

	return err == nil
}

func GetSepaByID(db bank.Database, SepaInfoID model.SepaInfoID) (model.FiatSepaInfo, error) {
	var result model.FiatSepaInfo

	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	err := gdb.Model(&model.FiatSepaInfo{}).
		Scopes(ScopeSepaInfoID(SepaInfoID)).
		First(&result).Error

	return result, err
}

func GetUserSepaIDList(db bank.Database, userID model.UserID) ([]model.SepaInfoID, error) {
	var result []model.SepaInfoID

	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return result, errors.New("Invalid appcontext.Database")
	}

	var list []*model.FiatSepaInfo
	err := gdb.Model(&model.FiatSepaInfo{}).
		Scopes(ScopeUserID(userID)).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertSepaInfoToIdList(list), err
}

func GetSepaByUserAndIban(db bank.Database, userID model.UserID, iban model.Iban) (model.FiatSepaInfo, error) {
	list, err := QuerySepaList(db, userID, iban)
	if err != nil {
		return model.FiatSepaInfo{}, err
	}

	if len(list) > 0 {
		return list[0], nil
	}
	return model.FiatSepaInfo{}, ErrSepaNotFound
}

func QuerySepaList(db bank.Database, userID model.UserID, iban model.Iban) ([]model.FiatSepaInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if gdb == nil {
		return nil, errors.New("Invalid appcontext.Database")
	}

	var filters []func(db *gorm.DB) *gorm.DB
	if userID == 0 {
		return nil, ErrInvalidUserID
	}

	filters = append(filters, ScopeUserID(userID))
	filters = append(filters, ScopeSepaInfoIban(iban))

	var list []*model.FiatSepaInfo
	err := gdb.Model(&model.FiatSepaInfo{}).
		Scopes(filters...).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return convertSepaInfoList(list), nil
}

func ScopeSepaInfoID(sepaInfoID model.SepaInfoID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqSepaInfoID(), sepaInfoID)
	}
}
func ScopeSepaInfoIban(iban model.Iban) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(reqSepaInfoIban(), iban)
	}
}

// Takes a slice of references to SepaInfo and returns a slice of SepaInfo values
func convertSepaInfoList(list []*model.FiatSepaInfo) []model.FiatSepaInfo {
	var result []model.FiatSepaInfo
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *curr)
	}

	return result[:]
}

// Takes a slice of references to SepaInfo and returns a slice of SepaInfoID
func convertSepaInfoToIdList(list []*model.FiatSepaInfo) []model.SepaInfoID {
	var result []model.SepaInfoID
	for _, curr := range list {
		if curr == nil {
			continue
		}
		result = append(result, *&curr.ID)
	}

	return result[:]
}

const (
	colSepaInfoIban  = "iban"
	colSepaInfoBic   = "bic"
	colSepaInfoLabel = "label"
)

func sepaInfoColumnNames() []string {
	return []string{
		colID,
		colUserID,
		colSepaInfoIban,
		colSepaInfoBic,
		colSepaInfoLabel,
	}
}

// zero allocation requests string for scope
func reqSepaInfoIban() string {
	var req [len(colSepaInfoIban) + len(reqEQ)]byte
	off := 0
	off += copy(req[off:], colSepaInfoIban)
	copy(req[off:], reqEQ)

	return string(req[:])
}
