package api

import (
	"git.condensat.tech/bank/database/model"
)

func Models() []model.Model {
	return []model.Model{
		new(model.User),
		new(model.UserRole),
		new(model.UserPGP),
		new(model.Credential),
		new(model.OAuth),
		new(model.OAuthData),
	}
}
