package commands

import (
	"context"

	"git.condensat.tech/bank/wallet/rpc"
)

const (
	AddressTypeBech32 = "bech32"
)

func GetNewAddress(ctx context.Context, rpcClient RpcClient, label, addressType string) (Address, error) {
	return GetNewAddressWithType(ctx, rpcClient, label, AddressTypeBech32)
}

func GetNewAddressWithType(ctx context.Context, rpcClient RpcClient, label, addressType string) (Address, error) {
	if rpcClient == nil {
		return "", ErrInvalidRPCClient
	}

	var address Address
	err := callCommand(rpcClient, CmdGetNewAddress, &address, label, addressType)
	if err != nil {
		return "", rpc.ErrRpcError
	}

	return address, nil
}
