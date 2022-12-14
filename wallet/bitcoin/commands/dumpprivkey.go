package commands

import (
	"context"
)

func DumpPrivkey(ctx context.Context, rpcClient RpcClient, address Address) (Address, error) {
	var result Address
	err := callCommand(rpcClient, CmdDumpPrivkey, &result, address)
	if err != nil {
		return "", err
	}

	return result, nil
}

func DumpPrivkeys(ctx context.Context, rpcClient RpcClient, addresses []Address) ([]Address, error) {
	var result []Address
	for _, address := range addresses {
		privkey, err := DumpPrivkey(ctx, rpcClient, address)
		if err != nil {
			result = append(result, Address(""))
			continue
		}
		result = append(result, privkey)
	}

	return result, nil
}
