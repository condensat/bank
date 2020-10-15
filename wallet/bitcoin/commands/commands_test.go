package commands

import (
	"git.condensat.tech/bank/wallet/common"
	"git.condensat.tech/bank/wallet/rpc"
)

func testRpcClient(hostname string, port int) RpcClient {
	return rpc.New(rpc.Options{
		ServerOptions: common.ServerOptions{Protocol: "http", HostName: hostname, Port: port},
		User:          "condensat",
		Password:      "condensat",
	}).Client
}
