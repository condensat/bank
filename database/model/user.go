package model

type User struct {
	ID    uint64 `gorm:"primary_key"`
	Name  string `gorm:"size:64;unique;not null"`
	Email string `gorm:"size:256;unique;not null"`
}
