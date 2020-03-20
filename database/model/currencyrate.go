package model

import (
	"time"
)

type CurrencyRate struct {
	ID        uint64    `gorm:"primary_key"`
	Timestamp time.Time `gorm:"index;not null;type:timestamp"`
	Source    string    `gorm:"index;not null;type:varchar(16)"`
	Base      string    `gorm:"index;not null;type:varchar(16)"`
	Name      string    `gorm:"index;not null;type:varchar(16)"`
	Rate      float64   `gorm:"not null"`
}
