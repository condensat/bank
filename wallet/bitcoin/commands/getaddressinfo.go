package commands

import (
	"context"
)

func GetAddressInfo(ctx context.Context, rpcClient RpcClient, address Address) (AddressInfo, error) {
	var result AddressInfo
	err := callCommand(rpcClient, CmdGetAddressInfo, &result, address)
	if err != nil {
		return AddressInfo{}, err
	}

	return result, nil
}
