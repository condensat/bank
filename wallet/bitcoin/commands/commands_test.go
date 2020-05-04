package commands

import (
	"git.condensat.tech/bank"

	"git.condensat.tech/bank/wallet/rpc"
)

func testRpcClient(hostname string, port int) RpcClient {
	return rpc.New(rpc.Options{
		ServerOptions: bank.ServerOptions{Protocol: "http", HostName: hostname, Port: port},
		User:          "condensat",
		Password:      "condensat",
	}).Client
}
