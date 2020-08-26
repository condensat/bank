package commands

import (
	"context"
)

func BlindRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (Transaction, error) {
	var result Transaction
	err := callCommand(rpcClient, CmdBlindRawTransaction, &result, hex)
	if err != nil {
		return Transaction(""), err
	}

	return result, nil
}
