package database

import (
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidAssetInfo = errors.New("Invalid AssetInfo")
)

// AddOrUpdateAssetInfo
func AddOrUpdateAssetInfo(db bank.Database, entry model.AssetInfo) (model.AssetInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.AssetInfo{}, errors.New("Invalid appcontext.Database")
	}

	if !entry.Valid() {
		return model.AssetInfo{}, ErrInvalidAssetInfo
	}

	entry.LastUpdate = time.Now().UTC().Truncate(time.Second)

	var result model.AssetInfo
	err := gdb.
		Where(model.AssetInfo{
			AssetID: entry.AssetID,
		}).
		Assign(entry).
		FirstOrCreate(&result).Error

	return result, err
}

// GetAssetInfo
func GetAssetInfo(db bank.Database, assetID model.AssetID) (model.AssetInfo, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.AssetInfo{}, errors.New("Invalid appcontext.Database")
	}

	if assetID == 0 {
		return model.AssetInfo{}, ErrInvalidAssetID
	}

	var result model.AssetInfo
	err := gdb.
		Where(&model.AssetInfo{AssetID: assetID}).
		First(&result).Error
	if err != nil {
		return model.AssetInfo{}, err
	}

	return result, nil
}
