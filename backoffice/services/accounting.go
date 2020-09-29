package services

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
)

type CurrencyBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Locked   float64 `json:"locked"`
}

type AccountingStatus struct {
	Count    int               `json:"count"`
	Active   int               `json:"active"`
	Balances []CurrencyBalance `json:"balances"`
}

func FetchAccountingStatus(ctx context.Context) (AccountingStatus, error) {
	db := appcontext.Database(ctx)

	accountsInfo, err := database.AccountsInfos(db)
	if err != nil {
		return AccountingStatus{}, err
	}

	var balances []CurrencyBalance
	for _, account := range accountsInfo.Accounts {
		balances = append(balances, CurrencyBalance{
			Currency: account.CurrencyName,
			Balance:  account.Balance,
			Locked:   account.TotalLocked,
		})
	}

	return AccountingStatus{
		Count:    accountsInfo.Count,
		Active:   accountsInfo.Active,
		Balances: balances,
	}, nil
}
