package model

type Credential struct {
	UserID       UserID `gorm:"unique_index"`
	LoginHash    Base58 `gorm:"size:64;not null;index"`
	PasswordHash Base58 `gorm:"size:64;not null;index"`
	TOTPSecret   String `gorm:"size:64;not null"`
}
