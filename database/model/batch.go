package model

import (
	"time"
)

type BatchID ID
type BatchData String

type Batch struct {
	ID        BatchID   `gorm:"primary_key"`
	Timestamp time.Time `gorm:"index;not null;type:timestamp"`   // Creation timestamp
	Data      BatchData `gorm:"type:blob;not null;default:'{}'"` // Batch data
}
