package commands

import (
	"context"
)

func ImportBlindingKey(ctx context.Context, rpcClient RpcClient, address Address, blindingKey BlindingKey) error {
	var noResult GenericJson
	err := callCommand(rpcClient, CmdImportBlindingKey, &noResult, address, blindingKey)
	if err != nil {
		return err
	}

	return nil
}
