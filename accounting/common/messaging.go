package common

const (
	chanPrefix = "Condensat.Accounting."

	CryptoFetchPendingWithdrawSubject = chanPrefix + "Crypto.FetchPendingWithdraw"
	CryptoValidateWithdrawSubject     = chanPrefix + "Crypto.ValidateWithdraw"

	FiatWithdrawSubject             = chanPrefix + "Fiat.Withdraw"
	FiatFetchPendingWithdrawSubject = chanPrefix + "Fiat.FetchPendingWithdraw"
	FiatValidateWithdrawSubject     = chanPrefix + "Fiat.ValidateWithdraw"
	FiatDepositSubject              = chanPrefix + "Fiat.Deposit"

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
	AccountTransferSubject  = chanPrefix + "Account.Transfer"

	AccountTransferWithdrawCryptoSubject = chanPrefix + "Account.TransferWithdrawCrypto"
	AccountTransferWithdrawFiatSubject   = chanPrefix + "Account.TransferWithdrawFiat"

	BatchWithdrawListSubject   = chanPrefix + "BatchWithdraw.List"
	BatchWithdrawUpdateSubject = chanPrefix + "BatchWithdraw.Update"

	UserWithdrawListSubject = chanPrefix + "User.Withdraw.List"

	CancelWithdrawSubject = chanPrefix + "Withdraw.Cancel"
)
