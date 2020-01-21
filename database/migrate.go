package database

import "git.condensat.tech/bank/database/model"

func (p *Database) Migrate(models []model.Model) error {
	var interfaces []interface{}
	for _, model := range models {
		interfaces = append(interfaces, model)
	}
	return p.db.AutoMigrate(
		interfaces...,
	).Error
}
