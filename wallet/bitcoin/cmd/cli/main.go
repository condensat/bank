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

func RawReissuanceElements(ctx context.Context, rpcClient commands.RpcClient, entropy, tokenID, txID string, tokenAmount float64) {
	var isAssetReissuance bool
	switch ctxValue := ctx.Value(KeyReissueAsset).(type) {
	case bool:
		isAssetReissuance = ctxValue
	}

	switch {
	case isAssetReissuance == true:

		/*entropy := "fd34841f00d9a1feb2c900c40ffd85ae0f9ba8ba9f5f9e3ab53ef8bd1902a6c1"
		assetBlinding := "4b526d202e5863bf228c24ac47409907f00ec326fc807a42bf4787d9ded113cc"
		tokenID := "e1512d6001df36641dfa1db18315dc6fd399405963a0fedae4a042281df35e49"
		txID := "4d094df0af5305b896a5c183bbd718926884aa093a46d2ac0ad945d58a74e5b2"

		rpcClient := elementsRpcClient("elements-testnet", 28432)
		if rpcClient == nil {
			panic("Invalid rpcClient")
		}*/
		tokenUtxo, err := commands.ListUnspentWithAsset(ctx, rpcClient, nil, tokenID)
		if err != nil {
			panic(err)
		}
		assetBlinding := tokenUtxo[0].AssetBlinder

		//address := "AzpjyAMVQzUvJjpE3TbEZ1ATvTScw5poMndTvGjnnku8LtAjnh5q693iZWCvQjYKcFYJwKqY2njnvBM5"
		changeAddress := "el1qq2nwuqgmsqaef7fvaqwfaesvpu3mfrnsagpsqxycsly5qulakdwl44fjvjjguqnv46zxsecrfh03ps7ghrfcwtl9fjxukqeh7"
		tokenChangeAddress := "el1qq0sdn8lhwp2qsk6efgfy890tk94jdpnrtjec97kpq37d3tywggz94m0nctz4rd3wc4y45e9k34r237c57rkw5f2adtha2z4m5"

		hex, err := commands.CreateRawTransaction(ctx, rpcClient, []commands.UTXOInfo{
			{TxID: txID, Vout: 2},
		}, []commands.SpendInfo{
			//{Address: address, Amount: 0.003},
			{Address: tokenChangeAddress, Amount: tokenAmount},
		}, []commands.AssetInfo{
			//{Address: address, Asset: "b2e15d0d7a0c94e4e2ce0fe6e8691b9e451377f6e46e8045a86f7c4b5d4f0f23"},
			{Address: tokenChangeAddress, Asset: tokenID},
		})
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

		funded, err := commands.FundRawTransactionWithOptions(ctx,
			rpcClient,
			hex,
			commands.FundRawTransactionOptions{
				ChangeAddress:   changeAddress,
				IncludeWatching: true,
			},
		)
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
		reissued, err := commands.RawReissueAsset(
			ctx,
			rpcClient,
			commands.Transaction(funded.Hex),
			1000,
			"el1qqtg9yfl7v9954zz8m6cz2dhlpjsmhkqw7dejfxa5jxju7jkxylusjq6nyqxe0k6mdzkhzkuhk3l8mp480m4lvfvc88tap8k6v",
			entropy,
			assetBlinding,
			0,
		)
		if err != nil {
			panic(err)
		}
		log.Printf("RawReissueAsset OK:\n%s\n", reissued)

		blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(reissued))
		if err != nil {
			panic(err)
		}

		log.Printf("Blinded transaction OK:\n%s", blinded)

		toSign := blinded
		signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(toSign))
		if err != nil {
			panic(err)
		}
		if !signed.Complete {
			panic("SignRawTransactionWithWallet failed") //this is expected if using a watch-only wallet
		}
		log.Printf("Reissuance TX signed:\n%+v", signed.Hex)

		accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
		if err != nil {
			panic(err)
		}
		switch {
		case accepted.Allowed == true:
			log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)
		case accepted.Allowed == false:
			log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)
			log.Printf("Reject-reason: %+v\n", accepted.Reason)
		}
	default:
		fmt.Println("Asset reissuance is deactivated")
		return
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
