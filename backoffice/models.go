// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package backoffice

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func Models() []database.Model {
	return []database.Model{
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
