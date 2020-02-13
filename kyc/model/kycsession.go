package model

type KycSession struct {
	ID     uint64 `gorm:"primary_key"`
	UserID uint64 `gorm:"index"`
	Email  string `gorm:"size:256;index;not null"`
	Token  string `gorm:"size:256;index;not null"`
}
