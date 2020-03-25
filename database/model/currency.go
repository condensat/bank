package model

type CurrencyName String
type CurrencyAvailable ZeroInt

type Currency struct {
	Name      CurrencyName `gorm:"primary_key;type:varchar(16)"` // [PK] Currency
	Available ZeroInt      `gorm:"default:0;not null"`
}

func NewCurrency(name CurrencyName, available Int) Currency {
	if len(name) == 0 {
		return Currency{}
	}
	if available < 0 {
		return Currency{}
	}

	return Currency{
		Name:      name,
		Available: ZeroInt(&available),
	}
}

func (p *Currency) IsAvailable() bool {
	return len(p.Name) > 0 && p.Available != nil && *p.Available > 0
}
