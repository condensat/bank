package model

import "time"

type CryptoAddressID ID
type BlockID ID

const MemPoolBlockID = BlockID(1)

type CryptoAddress struct {
	ID               CryptoAddressID `gorm:"primary_key"`                    // [PK] CryptoAddress
	AccountID        AccountID       `gorm:"index;not null"`                 // [FK] Reference to Account table
	PublicAddress    String          `gorm:"unique_index;not null;size:128"` // CryptoAddress public key, non mutable
	Unconfidential   String          `gorm:"index;size:64"`                  // CryptoAddress unconfidential address, non mutable`
	Chain            String          `gorm:"index;not null;size:16"`         // CryptoAddress chain, non mutable
	CreationDate     *time.Time      `gorm:"index;not null"`                 // CryptoAddress creation date, non mutable
	FirstBlockId     BlockID         `gorm:"index;not null"`                 // Block height of the first transaction
	IgnoreAccounting bool            `gorm:"not null"`                       // This address is not for Accounting (change address)
}

func (p *CryptoAddress) IsUsed() bool {
	return p.FirstBlockId > 0
}

func (p *CryptoAddress) Confirmations(height BlockID) int {
	if !p.IsUsed() {
		return 0
	}
	if p.FirstBlockId <= MemPoolBlockID {
		return 0
	}
	if p.FirstBlockId > height {
		return 0
	}
	return 1 + int(height) - int(p.FirstBlockId)
}
