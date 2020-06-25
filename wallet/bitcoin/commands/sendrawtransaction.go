package commands

import (
	"context"
)

func SendRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (TxID, error) {
	var result TxID
	err := callCommand(rpcClient, CmdSendRawTransaction, &result, hex)
	if err != nil {
		return "", err
	}

	return result, nil
}
