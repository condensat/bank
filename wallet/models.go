package wallet

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	var result []model.Model
	result = append(result, database.CryptoAddressModel()...)
	result = append(result, database.OperationInfoModel()...)
	return result
}
