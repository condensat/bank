package common

const (
	chanPrefix = "Condensat.Accounting."

	CurrencyInfoSubject         = chanPrefix + "Currency.Info"
	CurrencyCreateSubject       = chanPrefix + "Currency.Create"
	CurrencyListSubject         = chanPrefix + "Currency.List"
	CurrencySetAvailableSubject = chanPrefix + "Currency.SetAvailable"

	AccountCreateSubject    = chanPrefix + "Account.Create"
	AccountInfoSubject      = chanPrefix + "Account.Info"
	AccountListSubject      = chanPrefix + "Account.List"
	AccountHistorySubject   = chanPrefix + "Account.History"
	AccountSetStatusSubject = chanPrefix + "Account.SetStatus"
	AccountOperationSubject = chanPrefix + "Account.Operation"
	AccountTransfertSubject = chanPrefix + "Account.Transfert"
)