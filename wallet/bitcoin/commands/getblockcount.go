package commands

import (
	"context"
	"errors"
)

var (
	ErrInvalidRPCClient = errors.New("Invalid RPC Client")
)

func GetBlockCount(ctx context.Context, rpcClient RpcClient) (int64, error) {
	if rpcClient == nil {
		return 0, ErrInvalidRPCClient
	}

	var blockount int64
	err := callCommand(rpcClient, CmdGetBlockCount, &blockount)
	if err != nil {
		return 0, err
	}

	return blockount, nil
}
