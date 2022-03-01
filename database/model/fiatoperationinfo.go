package model

import "time"

type FiatOperationInfoID ID
type FiatOperationStatus String

const (
	FiatOperationStatusNone     FiatOperationStatus = "none"
	FiatOperationStatusPending  FiatOperationStatus = "pending"
	FiatOperationStatusComplete FiatOperationStatus = "complete"
	FiatOperationStatusCanceled FiatOperationStatus = "canceled"
)

type FiatOperationInfo struct {
	ID                FiatOperationInfoID `gorm:"primary_key;"`                  // [PK] FiatOperationInfo
	SepaInfoID        SepaInfoID          `gorm:"index;not null"`                // [FK] Reference to sepaInfo table
	UserID            UserID              `gorm:"index;not null"`                // [FK] Reference to User table
	CurrencyName      CurrencyName        `gorm:"index;not null"`                // Currency used in the operation
	Amount            ZeroFloat           `gorm:"not null"`                      // Amount of the currency
	CreationTimestamp time.Time           `gorm:"index;not null;type:timestamp"` // When was the operation created
	UpdateTimestamp   time.Time           `gorm:"index;not null;type:timestamp"` // Timestamp of the last status change
	Type              OperationType       `gorm:"not null"`                      // Type
	Status            FiatOperationStatus `gorm:"not null"`                      // Status
}
