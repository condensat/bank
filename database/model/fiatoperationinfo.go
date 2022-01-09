package model

type FiatOperationInfo struct {
	Label  string        `gorm:"primary_key;"`   // [PK] A label to identify an IBAN
	IBAN   string        `gorm:"index;not null"` // [FK] IBAN
	BIC    string        `gorm:"index;not null"` // [FK] BIC
	Type   OperationType `gorm:"index;not null"` // [FK] Type
	Status string        `gorm:"index;not null"` // [FK] Status
}
