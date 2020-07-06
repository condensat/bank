package commands

import (
	"context"
)

func DecodeRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (RawTransaction, error) {
	var result RawTransaction
	err := callCommand(rpcClient, CmdDecodeRawTransaction, &result, hex)
	if err != nil {
		return RawTransaction{}, err
	}

	return result, nil
}
