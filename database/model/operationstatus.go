package model

import "time"

type OperationStatus struct {
	OperationInfoID OperationInfoID `gorm:"unique_index;not null"`           // [FK] Reference to OperationInfo table
	LastUpdate      time.Time       `gorm:"index;not null;type:timestamp"`   // Last update timestamp
	State           string          `gorm:"index;not null;type:varchar(16)"` // [enum] Operation synchroneous state (received, confirmed, settled)
	Accounted       string          `gorm:"index;not null;type:varchar(16)"` // Accounted state (see State)
}
