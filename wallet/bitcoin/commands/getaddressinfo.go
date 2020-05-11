package commands

import (
	"context"

	"git.condensat.tech/bank/wallet/rpc"
)

func GetAddressInfo(ctx context.Context, rpcClient RpcClient, address Address) (AddressInfo, error) {
	var result AddressInfo
	err := callCommand(rpcClient, CmdGetAddressInfo, &result, address)
	if err != nil {
		return AddressInfo{}, rpc.ErrRpcError
	}

	return result, nil
}
