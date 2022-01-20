package model

type FiatOperationInfoID ID
type FiatOperationStatus String

const (
	FiatOperationStatusUnauthorised FiatOperationStatus = "unauthorised"
	FiatOperationStatusPending      FiatOperationStatus = "pending"
	FiatOperationStatusComplete     FiatOperationStatus = "complete"
)

type FiatOperationInfo struct {
	ID           FiatOperationInfoID `gorm:"primary_key;"`   // [PK] FiatOperationInfo
	SepaInfoID   SepaInfoID          `gorm:"index;not null"` // [FK] Reference to sepaInfo table
	UserID       UserID              `gorm:"index;not null"` // [FK] Reference to User table
	CurrencyName CurrencyName        `gorm:"index;not null"` // Currency used in the operation
	Amount       ZeroFloat           `gorm:"not null"`       // Amount of the currency
	Type         OperationType       `gorm:"not null"`       // Type
	Status       FiatOperationStatus `gorm:"not null"`       // Status
}
