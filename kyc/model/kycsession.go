package model

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type KycSession struct {
	ID         uint64 `gorm:"primary_key"`
	UserID     uint64 `gorm:"index;unique_index:idx_user_synaps"`
	SynapsCode string `gorm:"index;unique_index:idx_user_synaps;size:64;not null"`
	Token      string `gorm:"index;size:36;not null"`
}

func AddKycSession(ctx context.Context, userID uint64, synapsCode string) (*KycSession, error) {
	db := appcontext.Database(ctx)

	switch db := db.DB().(type) {
	case *gorm.DB:

		var session KycSession
		err := db.
			Where(&KycSession{
				UserID:     userID,
				SynapsCode: synapsCode,
			}).
			Attrs(&KycSession{
				Token: uuid.New().String(),
			}).
			FirstOrCreate(&session).Error

		return &session, err

	default:
		return nil, database.ErrInvalidDatabase
	}
}
