package database

import (
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidSsmAddressID     = errors.New("Invalid SsmAddressID")
	ErrInvalidSsmPublicAddress = errors.New("Invalid PublicAddress ID")
)

func AddSsmAddress(db bank.Database, address model.SsmAddress, info model.SsmAddressInfo) (model.SsmAddressID, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, errors.New("Invalid appcontext.Database")
	}

	if !address.IsValid() {
		return 0, errors.New("Invalid address")
	}
	info.SsmAddressID = model.SsmAddressID(1)
	if !info.IsValid() {
		return 0, errors.New("Invalid address info")
	}

	result := address
	err := gdb.Create(&result).Error
	if err != nil {
		return 0, err
	}

	info.SsmAddressID = result.ID
	if !info.IsValid() {
		return model.SsmAddressID(0), errors.New("Invalid address info")
	}
	err = gdb.Create(&info).Error
	if err != nil {
		return 0, err
	}

	_, err = UpdateSsmAddressState(db, result.ID, model.SsmAddressStatusUnused)
	if err != nil {
		return 0, nil
	}

	return result.ID, nil
}

func CountSsmAddress(db bank.Database, chain model.SsmChain, fingerprint model.SsmFingerprint) (int, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, errors.New("Invalid appcontext.Database")
	}

	if len(chain) == 0 {
		return 0, errors.New("Invalid chain")
	}

	if len(fingerprint) == 0 {
		return 0, errors.New("Invalid fingerprint")
	}

	result := 0
	err := gdb.Model(&model.SsmAddressInfo{}).Where(&model.SsmAddressInfo{
		Chain:       chain,
		Fingerprint: fingerprint,
	}).Count(&result).Error
	if err != nil {
		return 0, err
	}

	return result, nil
}

func CountSsmAddressByState(db bank.Database, chain model.SsmChain, fingerprint model.SsmFingerprint, state model.SsmAddressStatus) (int, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, errors.New("Invalid appcontext.Database")
	}

	if len(chain) == 0 {
		return 0, errors.New("Invalid chain")
	}

	if len(fingerprint) == 0 {
		return 0, errors.New("Invalid fingerprint")
	}

	subQueryInfo := gdb.
		Model(&model.SsmAddressInfo{}).
		Where(&model.SsmAddressInfo{
			Chain:       chain,
			Fingerprint: fingerprint,
		}).
		SubQuery()

	subQueryLast := gdb.Model(&model.SsmAddressState{}).
		Select("MAX(id)").
		Group("ssm_address_id").
		SubQuery()

	result := 0
	err := gdb.Model(&model.SsmAddressState{}).
		Joins("JOIN (?) AS i ON i.ssm_address_id = ssm_address_state.ssm_address_id", subQueryInfo).
		Where("state = ?", state).
		Where("ssm_address_state.id IN (?)", subQueryLast).
		Count(&result).Error

	if err != nil {
		return 0, err
	}

	return result, nil
}

func GetSsmAddress(db bank.Database, addressID model.SsmAddressID) (model.SsmAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddress{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddress{}, ErrInvalidSsmAddressID
	}

	var result model.SsmAddress
	err := gdb.
		Where(&model.SsmAddress{ID: addressID}).
		First(&result).Error
	if err != nil {
		return model.SsmAddress{}, err
	}

	return result, nil
}

func GetSsmAddressInfo(db bank.Database, addressID model.SsmAddressID) (model.SsmAddressInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddressInfo{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddressInfo{}, ErrInvalidSwapID
	}

	var result model.SsmAddressInfo
	err := gdb.
		Where(&model.SsmAddressInfo{SsmAddressID: addressID}).
		First(&result).Error
	if err != nil {
		return model.SsmAddressInfo{}, err
	}

	return result, nil
}

func NextSsmAddressID(db bank.Database, chain model.SsmChain, fingerprint model.SsmFingerprint) (model.SsmAddressID, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, errors.New("Invalid appcontext.Database")
	}

	if len(chain) == 0 {
		return 0, errors.New("Invalid chain")
	}

	if len(fingerprint) == 0 {
		return 0, errors.New("Invalid fingerprint")
	}

	subQueryInfo := gdb.
		Model(&model.SsmAddressInfo{}).
		Where(&model.SsmAddressInfo{
			Chain:       chain,
			Fingerprint: fingerprint,
		}).
		SubQuery()

	subQueryLast := gdb.Model(&model.SsmAddressState{}).
		Select("MAX(id)").
		Group("ssm_address_id").
		SubQuery()

	result := model.SsmAddressState{}
	err := gdb.Model(&model.SsmAddressState{}).
		Joins("JOIN (?) AS i ON i.ssm_address_id = ssm_address_state.ssm_address_id", subQueryInfo).
		Where("ssm_address_state.id IN (?)", subQueryLast).
		Where("state = ?", model.SsmAddressStatusUnused).
		First(&result).Error
	if err != nil {
		return 0, err
	}

	return result.SsmAddressID, nil
}

func GetSsmAddressByPublicAddress(db bank.Database, publicAddress model.SsmPublicAddress) (model.SsmAddress, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddress{}, errors.New("Invalid appcontext.Database")
	}

	if len(publicAddress) == 0 {
		return model.SsmAddress{}, ErrInvalidCryptoAddressID
	}

	var result model.SsmAddress
	err := gdb.
		Where(&model.SsmAddress{PublicAddress: publicAddress}).
		First(&result).Error
	if err != nil {
		return model.SsmAddress{}, err
	}

	return result, nil
}

func GetSsmAddressState(db bank.Database, addressID model.SsmAddressID) (model.SsmAddressState, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddressState{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddressState{}, ErrInvalidSsmAddressID
	}

	var result model.SsmAddressState
	err := gdb.
		Where(&model.SsmAddressState{SsmAddressID: addressID}).
		Last(&result).Error
	if err != nil {
		return model.SsmAddressState{}, err
	}

	return result, nil
}

func UpdateSsmAddressState(db bank.Database, addressID model.SsmAddressID, status model.SsmAddressStatus) (model.SsmAddressState, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.SsmAddressState{}, errors.New("Invalid appcontext.Database")
	}

	if addressID == 0 {
		return model.SsmAddressState{}, ErrInvalidSsmAddressID
	}
	if len(status) == 0 {
		return model.SsmAddressState{}, ErrInvalidSsmAddressID
	}

	result := model.SsmAddressState{
		SsmAddressID: addressID,
		Timestamp:    time.Now().UTC().Truncate(time.Second),
		State:        status,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.SsmAddressState{}, err
	}

	return result, nil
}
