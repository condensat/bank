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

func UserModel() []model.Model {
	return []model.Model{
		model.Model(new(model.User)),
		model.Model(new(model.UserRole)),
		model.Model(new(model.UserPGP)),
	}
}

func AccountModel() []model.Model {
	return append(UserModel(), []model.Model{
		model.Model(new(model.Currency)),
		model.Model(new(model.Account)),
	}...)
}

func AccountStateModel() []model.Model {
	return append(AccountModel(), new(model.AccountState))
}

func AccountOperationModel() []model.Model {
	return append(AccountStateModel(), new(model.AccountOperation))
}

func CurrencyModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Currency)),
		model.Model(new(model.CurrencyRate)),
	}
}

func CryptoAddressModel() []model.Model {
	return []model.Model{
		model.Model(new(model.CryptoAddress)),
	}
}

func SsmAddressModel() []model.Model {
	return []model.Model{
		model.Model(new(model.SsmAddress)),
		model.Model(new(model.SsmAddressInfo)),
		model.Model(new(model.SsmAddressState)),
	}
}

func OperationInfoModel() []model.Model {
	return append(CryptoAddressModel(), []model.Model{
		model.Model(new(model.OperationInfo)),
		model.Model(new(model.OperationStatus)),
	}...)
}

func AssetModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Asset)),
		model.Model(new(model.AssetInfo)),
		model.Model(new(model.AssetIcon)),
	}
}

func SwapModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Swap)),
		model.Model(new(model.SwapInfo)),
	}
}

func WithdrawModel() []model.Model {
	return append(AccountOperationModel(), []model.Model{
		model.Model(new(model.Withdraw)),
		model.Model(new(model.WithdrawInfo)),
		model.Model(new(model.WithdrawTarget)),
		model.Model(new(model.Fee)),
		model.Model(new(model.FeeInfo)),
		model.Model(new(model.Batch)),
		model.Model(new(model.BatchInfo)),
		model.Model(new(model.BatchWithdraw)),
		model.Model(new(model.FiatOperationInfo)),
		model.Model(new(model.FiatSepaInfo)),
	}...)
}

func FeeModel() []model.Model {
	return []model.Model{
		model.Model(new(model.Fee)),
		model.Model(new(model.FeeInfo)),
	}
}
