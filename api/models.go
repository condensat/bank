package api

import (
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	return []model.Model{
		new(model.User),
		new(model.Credential),
		new(model.OAuth),
		new(model.OAuthData),
		new(model.Asset),
		new(model.AssetInfo),
		new(model.AssetIcon),
	}
}
