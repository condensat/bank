package commands

import (
	"context"
	"fmt"
)

// RawIssueAssetWithAsset This is the minimal number of arguments you need to pass to issue an asset
func RawIssueAssetWithAsset(ctx context.Context, rpcClient RpcClient, hex Transaction, assetAmount float64, assetAddress string) (IssuedTransaction, error) {
	return rawIssueAssetWithOptions(ctx, rpcClient, hex, RawIssueAssetOptions{
		AssetAmount:  assetAmount,
		AssetAddress: assetAddress,
		Blind:        true, //we suppose that we always want to blind issuance, maybe we should change it though
	})
}

func RawIssueAssetWithToken(ctx context.Context, rpcClient RpcClient, hex Transaction, assetAmount, tokenAmount float64, assetAddress, tokenAddress string) (IssuedTransaction, error) {
	return rawIssueAssetWithOptions(ctx, rpcClient, hex, RawIssueAssetOptions{
		AssetAmount:  assetAmount,
		AssetAddress: assetAddress,
		TokenAmount:  tokenAmount,
		TokenAddress: tokenAddress,
		Blind:        true,
	})
}

// RawIssueAssetWithContract and with token, maybe we could have another case with contractHash and without Token?
func RawIssueAssetWithContract(ctx context.Context, rpcClient RpcClient, hex Transaction, assetAmount, tokenAmount float64, assetAddress, tokenAddress, contractHash string) (IssuedTransaction, error) {
	return rawIssueAssetWithOptions(ctx, rpcClient, hex, RawIssueAssetOptions{
		AssetAmount:  assetAmount,
		AssetAddress: assetAddress,
		TokenAmount:  tokenAmount,
		TokenAddress: tokenAddress,
		ContractHash: contractHash, //this is 64B long
		Blind:        true,
	})
}

func rawIssueAssetWithOptions(ctx context.Context, rpcClient RpcClient, hex Transaction, options RawIssueAssetOptions) (IssuedTransaction, error) {
	var result []IssuedTransaction
	var data []interface{}
	data = append(data, options)
	fmt.Printf("Options to rawissueasset are %+v\n", data)
	err := callCommand(rpcClient, CmdRawIssueAsset, &result, hex, &data)
	if err != nil {
		return IssuedTransaction{}, err
	}
	return result[0], nil
}
