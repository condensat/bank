package model

import "time"

type ValidationID ID

type Validation struct {
	ID            ValidationID       `gorm:"primary_key"`
	UserID        UserID             `gorm:"index;not null"`                  // [FK] Reference to User table
	Timestamp     time.Time          `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	OperationType OperationType      `gorm:"index;not null;type:varchar(16)"` // [enum] Determine table for ReferenceID (deposit, withdraw, transfer, adjustment, none, other)
	ReferenceID   RefID              `gorm:"index;not null"`                  // [optional - FK] reference to another table, depending on operation type
	Type          WithdrawTargetType `gorm:"index;not null;size:16"`          // DataType [onchain, liquid, lightning, sepa, swift, card]
	Base          CurrencyName       `gorm:"type:varchar(16)"`
	Amount        ZeroFloat          `gorm:"default:0;not null"` // Notional amount used for validation
}
