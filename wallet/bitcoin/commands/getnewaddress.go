package commands

import (
	"context"
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
		return "", err
	}

	return address, nil
}
