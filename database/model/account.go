package model

type AccountID ID
type AccountName String

type Account struct {
	ID           AccountID    `gorm:"primary_key"`                     // [PK] Account
	UserID       UserID       `gorm:"index;not null"`                  // [FK] Reference to User table
	CurrencyName CurrencyName `gorm:"index;not null;type:varchar(16)"` // [FK] Reference to Currency table
	Name         AccountName  `gorm:"index;not null"`                  // [U] Unique Account name for User and Currency
}
