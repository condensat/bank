package model

import (
	"time"
)

// AssetIcon from https://assets.blockstream.info/icons.json
type AssetIcon struct {
	AssetID    AssetID   `gorm:"unique_index;not null"`         // [FK] Reference to Asset table
	LastUpdate time.Time `gorm:"index;not null;type:timestamp"` // Last update timestamp
	Data       []byte    `gorm:"type:blob;default:null"`        // Decoded data byte
}
