package monitor

import (
	"context"
	"errors"

	"git.condensat.tech/bank/appcontext"
	"github.com/jinzhu/gorm"
)

func AddProcessInfo(ctx context.Context, processInfo *ProcessInfo) error {
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return errors.New("Wrong database")
	}

	return db.Create(&processInfo).Error
}
