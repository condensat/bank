package commands

import (
	"context"
)

func GetRawTransaction(ctx context.Context, rpcClient RpcClient, txID TransactionID) (Transaction, error) {
	var result Transaction
	err := callCommand(rpcClient, CmdGetRawTransaction, &result, txID)
	if err != nil {
		return "", err
	}

	return result, nil
}
