package monitor

import (
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/monitor/common"
)

func Models() []model.Model {
	return []model.Model{
		model.Model(new(common.ProcessInfo)),
	}
}
