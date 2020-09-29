package backoffice

import (
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	return []model.Model{
		new(model.User),
		new(model.UserRole),

		new(model.Account),
		new(model.AccountState),
		new(model.AccountOperation),

		new(model.OperationStatus),
		new(model.Batch),
		new(model.BatchInfo),
		new(model.Withdraw),
		new(model.WithdrawInfo),

		new(model.SwapInfo),
		new(model.SwapInfo),
	}
}
