package commands

import (
	"context"
)

func ListLockUnspent(ctx context.Context, rpcClient RpcClient) ([]UTXOInfo, error) {
	var list []UTXOInfo
	err := callCommand(rpcClient, CmdListLockUnspent, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}
