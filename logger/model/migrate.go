package model

import (
	"github.com/jinzhu/gorm"
)

func Migrate(db *gorm.DB) error {
	// Automigrate all package models
	return db.AutoMigrate(
		new(LogEntry),
	).Error
}
