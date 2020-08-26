package model

import (
	"time"
)

type SsmAddressStateID ID
type SsmAddressStatus String

const (
	SsmAddressStatusUnused      = SsmAddressStatus("unused")
	SsmAddressStatusUsed        = SsmAddressStatus("used")
	SsmAddressStatusBlacklisted = SsmAddressStatus("blacklisted")
)

type SsmAddressState struct {
	ID           SsmAddressStateID `gorm:"primary_key;"`                  // [PK] SsmAddressState ID
	SsmAddressID SsmAddressID      `gorm:"index;not null"`                // [FK] Reference to SsmAddress table
	Timestamp    time.Time         `gorm:"index;not null;type:timestamp"` // Creation timestamp
	State        SsmAddressStatus  `gorm:"not null;size:64"`              // Ssm State [unused, used, blacklisted]
}
