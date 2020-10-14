package database

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/monitor/database/model"
)

func Models() []database.Model {
	return []database.Model{
		database.Model(new(model.ProcessInfo)),
	}
}
