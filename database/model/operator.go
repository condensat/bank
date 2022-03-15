package model

import (
	"time"
)

type OperatorID ID

type Operator struct {
	ID                 OperatorID         `gorm:"primary_key;unique_index:idx_id_previd;"` // [PK] Operator
	UserID             UserID             `gorm:"index;not null"`                          // [FK] Reference to User table; UserID of the operator
	AccountID          AccountID          `gorm:"index;not null"`                          // [FK] Reference to Account table
	Timestamp          time.Time          `gorm:"index;not null;type:timestamp"`           // Operation timestamp
	AccountOperationID AccountOperationID `gorm:"index;not null"`                          // [FK] Reference to AccountOperation table; not for every command
	Command            String             `gorm:"not null"`                                // Command called by operator
}
