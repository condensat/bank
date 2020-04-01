package common

const (
	chanPrefix = "Condensat.Accounting."

	CurrencyCreateSubject       = chanPrefix + "Currency.Create"
	CurrencyListSubject         = chanPrefix + "Currency.List"
	CurrencySetAvailableSubject = chanPrefix + "Currency.SetAvailable"

	AccountCreateSubject    = chanPrefix + "Account.Create"
	AccountListSubject      = chanPrefix + "Account.List"
	AccountHistorySubject   = chanPrefix + "Account.History"
	AccountSetStatusSubject = chanPrefix + "Account.SetStatus"
	AccountOperationSubject = chanPrefix + "Account.Operation"
)
