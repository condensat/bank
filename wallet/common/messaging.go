package common

const (
	chanPrefix = "Condensat.Wallet."

	CryptoAddressNextDepositSubject = chanPrefix + "CryptoAddress.NextDeposit"
	CryptoAddressNewDepositSubject  = chanPrefix + "CryptoAddress.NewDeposit"
	AddressInfoSubject              = chanPrefix + "CryptoAddress.AddressInfo"

	WalletStatusSubject = chanPrefix + "WalletStatus"
	WalletListSubject   = chanPrefix + "WalletList"

	AssetListIssuancesSubject = chanPrefix + "Asset.ListIssuances"
	AssetIssuanceSubject      = chanPrefix + "Asset.Issuance"
	AssetReissuanceSubject    = chanPrefix + "Asset.Reissuance"
	AssetBurnSubject          = chanPrefix + "Asset.Burn"
)
