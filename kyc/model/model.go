package model

import (
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	return []model.Model{
		new(KycSession),
	}
}
