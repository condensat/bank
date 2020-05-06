package wallet

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	var result []model.Model
	result = append(result, database.CryptoAddressModel()...)
	result = append(result, database.OperationInfoModel()...)
	result = append(result, new(model.Asset))
	return result
}
