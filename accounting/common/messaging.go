package common

const (
	chanPrefix = "Condensat.Accounting."

	CurrencyCreateSubject = chanPrefix + "Currency.Create"
	CurrencyListSubject   = chanPrefix + "Currency.List"

	AccountCreateSubject  = chanPrefix + "Account.Create"
	AccountListSubject    = chanPrefix + "Account.List"
	AccountHistorySubject = chanPrefix + "Account.History"
)
