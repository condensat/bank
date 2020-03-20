package model

type Account struct {
	ID           uint64 `gorm:"primary_key"`                     // [PK] Account
	UserID       uint64 `gorm:"index;not null"`                  // [FK] Reference to User table
	CurrencyName string `gorm:"index;not null;type:varchar(16)"` // [FK] Reference to Currency table
}
