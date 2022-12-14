package database

import (
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
	"github.com/jinzhu/gorm"
)

const (
	MaxAssetIconDataLen = (256 << 10) - 1
)

var (
	ErrInvalidAssetIcon = errors.New("Invalid AssetIcon")
	ErrAssetIconToLarge = errors.New("AssetIcon Too Large")
)

// AddOrUpdateAssetIcon
func AddOrUpdateAssetIcon(db bank.Database, entry model.AssetIcon) (model.AssetIcon, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.AssetIcon{}, errors.New("Invalid appcontext.Database")
	}

	if entry.AssetID == 0 {
		return model.AssetIcon{}, ErrInvalidAssetIcon
	}
	if len(entry.Data) == 0 {
		return model.AssetIcon{}, ErrInvalidAssetIcon
	}
	if len(entry.Data) > MaxAssetIconDataLen {
		return model.AssetIcon{}, ErrAssetIconToLarge
	}

	entry.LastUpdate = time.Now().UTC().Truncate(time.Second)

	var result model.AssetIcon
	err := gdb.
		Where(model.AssetIcon{
			AssetID: entry.AssetID,
		}).
		Assign(entry).
		FirstOrCreate(&result).Error

	return result, err
}

// GetAssetIcon
func GetAssetIcon(db bank.Database, assetID model.AssetID) (model.AssetIcon, error) {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		return model.AssetIcon{}, errors.New("Invalid appcontext.Database")
	}

	if assetID == 0 {
		return model.AssetIcon{}, ErrInvalidAssetID
	}

	var result model.AssetIcon
	err := gdb.
		Where(&model.AssetIcon{AssetID: assetID}).
		First(&result).Error
	if err != nil {
		return model.AssetIcon{}, err
	}

	return result, nil
}
