package model

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type KycSession struct {
	ID     uint64 `gorm:"primary_key"`
	UserID uint64 `gorm:"index;unique_index:idx_user_email"`
	Email  string `gorm:"size:256;index;unique_index:idx_user_email;not null"`
	Token  string `gorm:"size:36;index;not null"`
}

func AddKycSession(ctx context.Context, userID uint64, email string) (*KycSession, error) {
	db := appcontext.Database(ctx)

	switch db := db.DB().(type) {
	case *gorm.DB:

		var session KycSession
		err := db.
			Where(&KycSession{
				UserID: userID,
				Email:  email,
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
