package commands

import (
	"context"

	"git.condensat.tech/bank/wallet/rpc"
)

func LockUnspent(ctx context.Context, rpcClient RpcClient, unlock bool, utxos []UTXOInfo) (bool, error) {
	var success bool
	err := callCommand(rpcClient, CmdLockUnspent, &success, unlock, utxos)
	if err != nil {
		return false, rpc.ErrRpcError
	}

	return success, nil
}
