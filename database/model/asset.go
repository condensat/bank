package model

type AssetID ID
type AssetHash String

type Asset struct {
	ID           AssetID      `gorm:"primary_key;"`                                          // [PK] Asset
	CurrencyName CurrencyName `gorm:"index;unique_index:idx_currency_hash;not null;size:16"` // [FK] Currency, non mutable`
	Hash         AssetHash    `gorm:"index;unique_index:idx_currency_hash;not null;size:64"` //  Asset Hash
}
