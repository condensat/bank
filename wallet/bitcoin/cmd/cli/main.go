package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/wallet/rpc"

	"git.condensat.tech/bank/wallet/bitcoin/commands"

	dotenv "github.com/joho/godotenv"
)

const (
	KeyIssueAsset = "Key.IssueAsset"
)

func init() {
	_ = dotenv.Load()
}

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyIssueAsset, true)
	RawTransaction(ctx)
	RawTransactionElements(ctx)
}

func RawTransaction(ctx context.Context) {
	rpcClient := bitcoinRpcClient("bitcoin-testnet", 28332)
	if rpcClient == nil {
		panic("Invalid rpcClient")
	}
	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: "bcrt1qjlw9gfrqk0w2ljegl7vwzrt2rk7sst8d4hm7n9", Amount: 0.003},
	}, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	rawTx, err := commands.DecodeRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		panic(err)
	}
	decoded, err := commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		panic(err)
	}
	log.Printf("DecodeRawTransaction: %+v\n", decoded)

	funded, err := commands.FundRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		panic(err)
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	rawTx, err = commands.DecodeRawTransaction(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		panic(err)
	}
	decoded, err = commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		panic(err)
	}
	log.Printf("FundRawTransaction Hex: %+v\n", decoded)

	addressMap := make(map[commands.Address]commands.Address)
	for _, in := range decoded.Vin {

		txInfo, err := commands.GetTransaction(ctx, rpcClient, in.Txid, true)
		if err != nil {
			panic(err)
		}

		addressMap[txInfo.Address] = txInfo.Address
		for _, d := range txInfo.Details {
			address := commands.Address(d.Address)
			addressMap[address] = address
		}
	}

	signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		panic(err)
	}
	if !signed.Complete {
		panic("SignRawTransactionWithWallet failed")
	}
	log.Printf("Signed transaction is: %+v\n", signed.Hex)

	accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
	if err != nil {
		panic(err)
	}
	log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)

}

func RawTransactionElements(ctx context.Context) {
	var isAssetIssuance bool
	switch ctxValue := ctx.Value(KeyIssueAsset).(type) {
	case bool:
		isAssetIssuance = ctxValue
	}
	rpcClient := elementsRpcClient("elements-testnet", 18432)
	if rpcClient == nil {
		panic("Invalid rpcClient")
	}

	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: "AzpjyAMVQzUvJjpE3TbEZ1ATvTScw5poMndTvGjnnku8LtAjnh5q693iZWCvQjYKcFYJwKqY2njnvBM5", Amount: 0.003},
	}, nil)
	if err != nil {
		panic(err)
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	rawTx, err := commands.DecodeRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		panic(err)
	}
	decoded, err := commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		panic(err)
	}
	log.Printf("DecodeRawTransaction: %+v\n", decoded)

	funded, err := commands.FundRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		panic(err)
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	rawTx, err = commands.DecodeRawTransaction(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		panic(err)
	}
	decoded, err = commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		panic(err)
	}
	log.Printf("FundRawTransaction Hex: %+v\n", decoded)

	addressMap := make(map[commands.Address]commands.Address)
	for _, in := range decoded.Vin {

		txInfo, err := commands.GetTransaction(ctx, rpcClient, in.Txid, true)
		if err != nil {
			panic(err)
		}

		addressMap[txInfo.Address] = txInfo.Address
		for _, d := range txInfo.Details {
			address := commands.Address(d.Address)
			addressMap[address] = address
		}
	}

	switch {
	case isAssetIssuance == true:

		tx := funded.Hex
		issuedWithAsset, err := commands.RawIssueAssetWithAsset(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			1000.00000001,
			"AzpuKUXvtnbf5uGQrpvDCYrnPJPCaYbRCLk39o73zJxq4od1u9jbcokao9fvFJQp4D9iSMUpLiuknSmR",
		)
		if err != nil {
			panic(err)
		}
		log.Printf("RawIssueAssetWithAsset OK, issued asset is %s\n", issuedWithAsset.Asset)

		issuedWithToken, err := commands.RawIssueAssetWithToken(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			1000.00000001,
			0.00000001,
			"AzpuKUXvtnbf5uGQrpvDCYrnPJPCaYbRCLk39o73zJxq4od1u9jbcokao9fvFJQp4D9iSMUpLiuknSmR",
			"el1qq0kw2e8g4yknjc0uggucchpucpcut47cl2xeymmqylt5vw6pt9dgl5pf9je6yxjcx7tw6ewmfv5hj54009d2yzgyqyyuulxlf",
		)
		if err != nil {
			panic(err)
		}
		log.Printf("RawIssueAssetWithToken OK, issued asset is %s\n", issuedWithToken.Asset)

		issuedWithContract, err := commands.RawIssueAssetWithContract(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			1000.00000001,
			"AzpuKUXvtnbf5uGQrpvDCYrnPJPCaYbRCLk39o73zJxq4od1u9jbcokao9fvFJQp4D9iSMUpLiuknSmR",
			"7F6475E61926B63C190CEBAB3470531EAB53B5481CBF960C3EE3164CA71E816B",
		)
		if err != nil {
			panic(err)
		}
		log.Printf("RawIssueAssetWithContract OK, issued asset is %s\n", issuedWithContract.Asset)

		issuedWithTokenWithContract, err := commands.RawIssueAssetWithTokenWithContract(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			1000.00000001,
			0.00000001,
			"AzpuKUXvtnbf5uGQrpvDCYrnPJPCaYbRCLk39o73zJxq4od1u9jbcokao9fvFJQp4D9iSMUpLiuknSmR",
			"el1qq0kw2e8g4yknjc0uggucchpucpcut47cl2xeymmqylt5vw6pt9dgl5pf9je6yxjcx7tw6ewmfv5hj54009d2yzgyqyyuulxlf",
			"7F6475E61926B63C190CEBAB3470531EAB53B5481CBF960C3EE3164CA71E816B",
		)
		if err != nil {
			panic(err)
		}

		log.Printf("RawIssueAssetWithTokenWithContract OK, issued asset is %s\n", issuedWithTokenWithContract.Asset)

		toSign := issuedWithTokenWithContract.Hex
		blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(toSign))
		if err != nil {
			panic(err)
		}

		log.Printf("Blinded transaction OK\n")
		log.Printf("issuedWithTokenWithContract to sign:\n%+v", blinded)

		signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(blinded))
		if err != nil {
			panic(err)
		}
		if !signed.Complete {
			panic("SignRawTransactionWithWallet failed")
		}
		log.Printf("issuedWithTokenWithContract signed:\n%+v", signed.Hex)

		accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
		if err != nil {
			panic(err)
		}
		log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)

	default:
		tx := funded.Hex
		blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(tx))
		if err != nil {
			panic(err)
		}

		log.Printf("Blinded transaction OK\n")

		signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(blinded))
		if err != nil {
			panic(err)
		}
		if !signed.Complete {
			panic("SignRawTransactionWithWallet failed")
		}

		accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
		if err != nil {
			panic(err)
		}
		log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)

	}

}

func bitcoinRpcClient(hostname string, port int) commands.RpcClient {
	password := os.Getenv("BITCOIN_TESTNET_PASSWORD")
	return rpc.New(rpc.Options{
		ServerOptions: bank.ServerOptions{Protocol: "http", HostName: hostname, Port: port},
		User:          "bank-wallet",
		Password:      password,
	}).Client
}

func elementsRpcClient(hostname string, port int) commands.RpcClient {
	password := os.Getenv("ELEMENTS_TESTNET_PASSWORD")
	return rpc.New(rpc.Options{
		ServerOptions: bank.ServerOptions{Protocol: "http", HostName: hostname, Port: port},
		User:          "bank-wallet",
		Password:      password,
	}).Client
}
