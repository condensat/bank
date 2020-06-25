package commands

type Command string

const (
	CmdGetBlockCount   = Command("getblockcount")
	CmdGetNewAddress   = Command("getnewaddress")
	CmdListUnspent     = Command("listunspent")
	CmdLockUnspent     = Command("lockunspent")
	CmdListLockUnspent = Command("listlockunspent")
	CmdGetTransaction  = Command("gettransaction")
	CmdGetAddressInfo  = Command("getaddressinfo")
	CmdSendMany        = Command("sendmany")

	CmdCreateRawTransaction = Command("createrawtransaction")
)
