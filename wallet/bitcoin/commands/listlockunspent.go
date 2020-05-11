package commands

import (
	"context"

	"git.condensat.tech/bank/wallet/rpc"
)

func ListLockUnspent(ctx context.Context, rpcClient RpcClient) ([]UTXOInfo, error) {
	var list []UTXOInfo
	err := callCommand(rpcClient, CmdListLockUnspent, &list)
	if err != nil {
		return nil, rpc.ErrRpcError
	}

	return list, nil
}
