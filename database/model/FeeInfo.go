package model

import (
	"git.condensat.tech/bank/utils"
)

const DefaultFeeRate = Float(0.001) // 0.1%

type FeeInfo struct {
	Currency CurrencyName `gorm:"primary_key"`            // [PK] Related currency
	Minimum  Float        `gorm:"default:0.0;not null"`   // Minimum Fee
	Rate     Float        `gorm:"default:0.001;not null"` // Percent Fee Rate (default 0.1%)
}

func (p *FeeInfo) IsValid() bool {
	return len(p.Currency) > 0 &&
		p.Minimum >= 0.0 &&
		p.Rate >= 0.0
}

func (p *FeeInfo) Compute(amount Float) Float {
	if !p.IsValid() {
		return 0.0
	}
	if amount <= 0.0 {
		return 0.0
	}

	fees := amount * p.Rate
	if fees < p.Minimum {
		fees = p.Minimum
	}
	fees = Float(utils.ToFixed(float64(fees), utils.DatabaseFloatingPrecision))

	return Float(fees)
}
