package commands

import (
	"context"

	"git.condensat.tech/bank/wallet/rpc"
)

func ImportAddress(ctx context.Context, rpcClient RpcClient, address Address, label string, reindex bool) error {
	var noResult GenericJson
	err := callCommand(rpcClient, CmdImportAddress, &noResult, address, label, reindex)
	if err != nil {
		return rpc.ErrRpcError
	}

	return nil
}
