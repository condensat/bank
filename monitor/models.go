package monitor

import (
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	return []model.Model{
		model.Model(new(ProcessInfo)),
	}
}
