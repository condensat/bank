package commands

type Command string

const (
	CmdGetBlockCount       = Command("getblockcount")
	CmdGetNewAddress       = Command("getnewaddress")
	CmdListUnspent         = Command("listunspent")
	CmdLockUnspent         = Command("lockunspent")
	CmdListLockUnspent     = Command("listlockunspent")
	CmdGetTransaction      = Command("gettransaction")
	CmdGetRawTransaction   = Command("getrawtransaction")
	CmdGetAddressInfo      = Command("getaddressinfo")
	CmdImportAddress       = Command("importaddress")
	CmdImportPubKey        = Command("importpubkey")
	CmdImportBlindingKey   = Command("importblindingkey")
	CmdBlindRawTransaction = Command("blindrawtransaction")
	CmdSendMany            = Command("sendmany")

	CmdDumpPrivkey                  = Command("dumpprivkey")
	CmdCreateRawTransaction         = Command("createrawtransaction")
	CmdDecodeRawTransaction         = Command("decoderawtransaction")
	CmdFundRawTransaction           = Command("fundrawtransaction")
	CmdSignRawTransactionWithKey    = Command("signrawtransactionwithkey")
	CmdSignRawTransactionWithWallet = Command("signrawtransactionwithwallet")
	CmdSendRawTransaction           = Command("sendrawtransaction")
)
