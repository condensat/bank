package model

import (
	"time"
)

type OperationInfoID ID
type TxID String
type Vout ID

type OperationInfo struct {
	ID              OperationInfoID `gorm:"primary_key;"`                  // [PK] OperationInfo
	CryptoAddressID CryptoAddressID `gorm:"index;not null"`                // [FK] Reference to CryptoAddress table
	AssetID         AssetID         `gorm:"default:0;not null"`            // [FK] Reference to Asset table, optional
	Timestamp       time.Time       `gorm:"index;not null;type:timestamp"` // Creation timestamp
	TxID            TxID            `gorm:"unique_index;not null;size:64"` // Transaction ID
	Vout            Vout            `gorm:"default:0"`                     // Transaction output ID
	Amount          Float           `gorm:"default:0.0;not null"`          // Operation amount (GTE 0.0)
	Data            String          `gorm:"type:json;not null"`            // Specific operation json data
}
