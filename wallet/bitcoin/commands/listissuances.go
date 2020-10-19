package commands

import "context"

func ListIssuances(ctx context.Context, rpcClient RpcClient, asset AssetID) ([]ListIssuancesInfo, error) {
	var result []ListIssuancesInfo

	err := callCommand(rpcClient, CmdListIssuances, &result, asset)
	if err != nil {
		return nil, err
	}
	return result, nil
}
