package model

import (
	"git.condensat.tech/bank/utils"
)

func ToFixedFloat(value Float) Float {
	fixed := utils.ToFixed(float64(value), utils.DatabaseFloatingPrecision)
	return Float(fixed)
}
