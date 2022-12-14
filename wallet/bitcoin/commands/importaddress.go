package commands

import (
	"context"
)

func ImportAddress(ctx context.Context, rpcClient RpcClient, address Address, label string, reindex bool) error {
	var noResult GenericJson
	err := callCommand(rpcClient, CmdImportAddress, &noResult, address, label, reindex)
	if err != nil {
		return err
	}

	return nil
}
