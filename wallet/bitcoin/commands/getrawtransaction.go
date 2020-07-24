package commands

import (
	"context"

	"git.condensat.tech/bank/wallet/rpc"
)

func GetRawTransaction(ctx context.Context, rpcClient RpcClient, txID TransactionID) (Transaction, error) {
	var result Transaction
	err := callCommand(rpcClient, CmdGetRawTransaction, &result, txID)
	if err != nil {
		return "", rpc.ErrRpcError
	}

	return result, nil
}
