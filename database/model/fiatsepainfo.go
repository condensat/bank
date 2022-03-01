package model

type SepaInfoID ID
type Iban String
type Bic String

// TODO: define types for iban/bic and maybe sanity checks (length, structure...)

type FiatSepaInfo struct {
	ID     SepaInfoID `gorm:"primary_key;"`   // [PK] SepaInfoID
	UserID UserID     `gorm:"index;not null"` // [FK] reference to User table
	IBAN   Iban       `gorm:"not null"`       // IBAN
	BIC    Bic        `gorm:"not null"`       // BIC
	Label  String     `gorm:"not null"`       // A label to identify an IBAN
}
