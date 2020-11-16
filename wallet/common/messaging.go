package common

const (
	chanPrefix = "Condensat.Wallet."

	CryptoAddressNextDepositSubject = chanPrefix + "CryptoAddress.NextDeposit"
	CryptoAddressNewDepositSubject  = chanPrefix + "CryptoAddress.NewDeposit"
	AddressInfoSubject              = chanPrefix + "CryptoAddress.AddressInfo"

	WalletStatusSubject = chanPrefix + "WalletStatus"
	WalletListSubject   = chanPrefix + "WalletList"

	AssetIssuanceSubject = chanPrefix + "Asset.Issuance"
)
