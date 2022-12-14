package commands

import (
	"context"
	"errors"
)

var (
	ErrInvalidRPCClient = errors.New("Invalid RPC Client")
)

func NewAddress(ctx context.Context, rpcClient RpcClient, chain, fingerprint, path string) (NewAddressResponse, error) {
	if rpcClient == nil {
		return NewAddressResponse{}, ErrInvalidRPCClient
	}

	var address NewAddressResponse
	err := callCommand(rpcClient, CmdNewAddress, &address, chain, fingerprint, path)
	if err != nil {
		return NewAddressResponse{}, err
	}

	return address, nil
}
