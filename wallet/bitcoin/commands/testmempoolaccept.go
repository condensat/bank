package commands

import (
	"context"
	"errors"
)

func TestMempoolAccept(ctx context.Context, rpcClient RpcClient, hex string) (MempoolAccept, error) {
	var result []MempoolAccept
	var data []interface{}
	data = append(data, []string{hex})
	err := callCommand(rpcClient, CmdTestMempoolAccept, &result, data)
	if err != nil {
		return MempoolAccept{}, err
	}
	if len(result) != 1 {
		return MempoolAccept{}, errors.New("invalid mempoolaccept result")
	}

	return result[0], nil
}
