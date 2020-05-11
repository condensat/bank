package tasks

const (
	PolicyAssetLiquid = "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d"
	TickerAssetLiquid = "LBTC"

	ConfirmedBlockCount   = 3 // number of confirmation to consider transaction complete
	UnconfirmedBlockCount = 6 // number of confirmation to continue fetching addressInfos

	AddressInfoMinConfirmation = 0
	AddressInfoMaxConfirmation = 9999
)
