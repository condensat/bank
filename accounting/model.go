package accounting

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	return database.AccountOperationModel()
}
