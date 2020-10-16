// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package api

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func Models() []database.Model {
	return []database.Model{
		new(model.User),
		new(model.UserRole),
		new(model.Credential),
		new(model.OAuth),
		new(model.OAuthData),
		new(model.Asset),
		new(model.AssetInfo),
		new(model.AssetIcon),
		new(model.Swap),
		new(model.SwapInfo),
	}
}
