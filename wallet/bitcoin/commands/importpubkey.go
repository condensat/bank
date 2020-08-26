package commands

import (
	"context"
)

func ImportPubKey(ctx context.Context, rpcClient RpcClient, pubKey PubKey, label string, reindex bool) error {
	var noResult GenericJson
	err := callCommand(rpcClient, CmdImportPubKey, &noResult, pubKey, label, reindex)
	if err != nil {
		return err
	}

	return nil
}
