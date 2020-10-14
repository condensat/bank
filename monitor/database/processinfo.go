package database

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/monitor/database/model"

	"github.com/jinzhu/gorm"
)

func AddProcessInfo(db database.Context, processInfo *model.ProcessInfo) error {
	gdb := db.DB().(*gorm.DB)
	if db == nil {
		panic("Invalid db")
	}

	return gdb.Create(&processInfo).Error
}
