package database

import (
	bank "git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/monitor/database/model"
)

func Models() []bank.Model {
	return []bank.Model{
		bank.Model(new(model.ProcessInfo)),
	}
}
