// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/utils"
)

func ToFixedFloat(value Float) Float {
	fixed := utils.ToFixed(float64(value), database.DatabaseFloatingPrecision)
	return Float(fixed)
}
