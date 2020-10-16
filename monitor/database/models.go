// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package database

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/monitor/database/model"
)

func Models() []database.Model {
	return []database.Model{
		database.Model(new(model.ProcessInfo)),
	}
}
