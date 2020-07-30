package commands

import (
	"context"
)

func FundRawTransaction(ctx context.Context, rpcClient RpcClient, hex Transaction) (FundedTransaction, error) {
	var result FundedTransaction
	err := callCommand(rpcClient, CmdFundRawTransaction, &result, hex)
	if err != nil {
		return FundedTransaction{}, err
	}
	return result, nil
}
