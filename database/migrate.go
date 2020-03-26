package database

import (
	"git.condensat.tech/bank/database/model"
)

func (p *Database) Migrate(models []model.Model) error {
	var interfaces []interface{}
	for _, model := range models {
		interfaces = append(interfaces, model)
	}
	return p.db.AutoMigrate(
		interfaces...,
	).Error
}

func AccountModel() []model.Model {
	return []model.Model{
		model.Model(new(model.User)),
		model.Model(new(model.Currency)),
		model.Model(new(model.Account)),
	}
}

func CurrencyModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Currency)),
		model.Model(new(model.CurrencyRate)),
	}
}
