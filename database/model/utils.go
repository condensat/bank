package model

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/utils"
)

func ToFixedFloat(value Float) Float {
	fixed := utils.ToFixed(float64(value), database.DatabaseFloatingPrecision)
	return Float(fixed)
}
