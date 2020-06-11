package commands

import (
	"context"

	"git.condensat.tech/bank/wallet/rpc"
)

const (
	AddressInfoMinConfirmation = 0
	AddressInfoMaxConfirmation = 6
)

func ListUnspent(ctx context.Context, rpcClient RpcClient, filter []Address) ([]TransactionInfo, error) {
	return ListUnspentMinMaxAddresses(ctx, rpcClient, AddressInfoMinConfirmation, AddressInfoMaxConfirmation, filter)
}

func ListUnspentMinMaxAddresses(ctx context.Context, rpcClient RpcClient, minConf, maxConf int, filter []Address) ([]TransactionInfo, error) {
	list := make([]TransactionInfo, 0)
	err := callCommand(rpcClient, CmdListUnspent, &list, minConf, maxConf, filter)
	if err != nil {
		return nil, rpc.ErrRpcError
	}

	return list, nil
}