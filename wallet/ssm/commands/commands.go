package commands

type Command string

const (
	CmdNewAddress = Command("new_address")
)

type RpcClient interface {
	CallFor(out interface{}, method string, params ...interface{}) error
}

func callCommand(rpcClient RpcClient, command Command, out interface{}, params ...interface{}) error {
	return rpcClient.CallFor(out, string(command), params...)
}
