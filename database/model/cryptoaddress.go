package model

import "time"

type BlockID ID

type CryptoAddress struct {
	ID            ID         `gorm:"primary_key"`                   // [PK] CryptoAddress
	AccountID     AccountID  `gorm:"index;not null"`                // [FK] Reference to Account table
	PublicAddress String     `gorm:"unique_index;not null;size:64"` // CryptoAddress public key, non mutable
	CreationDate  *time.Time `gorm:"index;not null"`                // CryptoAddress creation date, non mutable
	FirstBlockId  BlockID    `gorm:"index;not null"`                // Block height of the first transaction
}
