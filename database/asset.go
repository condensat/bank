package database

import (
	"errors"
	"strings"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/utils"

	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidAssetID   = errors.New("Invalid AssetID")
	ErrInvalidAssetHash = errors.New("Invalid AssetHash")
)

func AddAsset(db bank.Database, assetHash model.AssetHash, currencyName model.CurrencyName) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, errors.New("Invalid appcontext.Database")
	}

	if len(assetHash) == 0 {
		return model.Asset{}, ErrInvalidAssetHash
	}

	if len(currencyName) == 0 {
		return model.Asset{}, ErrInvalidCurrencyName
	}

	result := model.Asset{
		Hash:         assetHash,
		CurrencyName: currencyName,
	}
	err := gdb.Create(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil

}

func AssetCount(db bank.Database) (int64, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return 0, errors.New("Invalid appcontext.Database")
	}

	var count int64
	err := gdb.Model(&model.Asset{}).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetAsset(db bank.Database, assetID model.AssetID) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, errors.New("Invalid appcontext.Database")
	}

	if assetID == 0 {
		return model.Asset{}, ErrInvalidAssetID
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{ID: assetID}).
		First(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil
}

func GetAssetByHash(db bank.Database, assetHash model.AssetHash) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, errors.New("Invalid appcontext.Database")
	}

	if len(assetHash) == 0 {
		return model.Asset{}, ErrInvalidAssetHash
	}

	if utils.ContainEllipsis(string(assetHash)) {
		assetHash = getFullAssetHash(gdb, assetHash)
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{Hash: assetHash}).
		First(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil
}

func GetAssetByCurrencyName(db bank.Database, currencyName model.CurrencyName) (model.Asset, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.Asset{}, errors.New("Invalid appcontext.Database")
	}

	if len(currencyName) == 0 {
		return model.Asset{}, ErrInvalidCurrencyName
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{CurrencyName: currencyName}).
		First(&result).Error
	if err != nil {
		return model.Asset{}, err
	}

	return result, nil
}

func AssetHashExists(db bank.Database, assetHash model.AssetHash) bool {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return false
	}

	if len(assetHash) == 0 {
		return false
	}

	var result model.Asset
	err := gdb.
		Where(&model.Asset{Hash: assetHash}).
		First(&result).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return false
	}

	return true
}

func getFullAssetHash(gdb *gorm.DB, assetHash model.AssetHash) model.AssetHash {
	tips := utils.SplitEllipsis(string(assetHash))
	if len(tips) != 2 {
		return assetHash
	}

	var result model.Asset
	err := gdb.
		Where("asset LIKE ?", strings.Join(tips, "%")).
		Where(&model.Asset{Hash: assetHash}).
		First(&result).Error
	if err != nil {
		return assetHash
	}

	return result.Hash
}
