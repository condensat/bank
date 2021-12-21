package model

type PgpPrivateKey String
type PgpPublicKey String

type UserPGP struct {
	UserID        UserID        `gorm:"unique_index:idx_user_role;index;not null"` // [FK] Reference to User table
	PgpPrivateKey PgpPrivateKey `gorm:"type:text;not null"`                        // PgpPrivateKey
	PublicKey     PgpPublicKey  `gorm:"type:text;not null"`                        // PgpPublicKey
}
