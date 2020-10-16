// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package wallet

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/query"
)

func Models() []database.Model {
	var result []database.Model
	result = append(result, query.CryptoAddressModel()...)
	result = append(result, query.SsmAddressModel()...)
	result = append(result, query.OperationInfoModel()...)
	result = append(result, query.AssetModel()...)
	return result
}
