package model

type Credential struct {
	UserID       uint64 `gorm:"unique_index"`
	LoginHash    string `gorm:"size:64;not null;index"`
	PasswordHash string `gorm:"size:64;not null;index"`
	TOTPSecret   string `gorm:"size:64;not null"`
}
