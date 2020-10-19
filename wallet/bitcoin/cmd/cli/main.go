package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/wallet/rpc"

	"git.condensat.tech/bank/wallet/bitcoin/commands"

	dotenv "github.com/joho/godotenv"
)

const (
	KeyIssueAsset   = "Key.IssueAsset"
	KeyReissueAsset = "Key.ReissueAsset"
)

func init() {
	_ = dotenv.Load()
}

func main() {
	var destAddress string
	var changeAddress string
	var assetAddress string
	var tokenAddress string
	var reissuedAsset string

	flag.StringVar(&destAddress, "dest", "", "Address to send L-BTC")
	flag.StringVar(&changeAddress, "change", "", "Address to send change")
	flag.StringVar(&assetAddress, "asset", "", "Address to send asset")
	flag.StringVar(&tokenAddress, "token", "", "Address to send token")
	flag.StringVar(&reissuedAsset, "reissue", "", "Asset to reissue")
	flag.Parse()

	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyIssueAsset, true)
	ctx = context.WithValue(ctx, KeyReissueAsset, true)
	RawTransaction(ctx)
	RawTransactionElements(ctx, destAddress, changeAddress, assetAddress, tokenAddress)
	RawReissuanceElements(ctx, reissuedAsset)
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
		return //abort the test if there's no bitcoind running
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

func RawTransactionElements(ctx context.Context, destAddress, changeAddress, assetAddress, tokenAddress string) {
	var isAssetIssuance bool
	switch ctxValue := ctx.Value(KeyIssueAsset).(type) {
	case bool:
		isAssetIssuance = ctxValue
	}
	rpcClient := elementsRpcClient("elements-testnet", 28433) // this may change
	if rpcClient == nil {
		panic("Invalid rpcClient")
	}

	// We create 2 LBTC outputs, which might be a bit unnecessary
	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: string(destAddress), Amount: 0.001},
	}, nil)
	if err != nil {
		panic(err)
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	rawTx, err := commands.DecodeRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		panic(err)
	}
	decoded, err := commands.ConvertToRawTransactionLiquid(rawTx)
	if err != nil {
		panic(err)
	}
	log.Printf("DecodeRawTransaction: %+v\n", decoded)

	funded, err := commands.FundRawTransactionWithOptions(ctx,
		rpcClient,
		hex,
		commands.FundRawTransactionOptions{
			ChangeAddress:   string(changeAddress),
			IncludeWatching: true,
		},
	)
	if err != nil {
		panic(err)
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	// I don't have usage for that here, but it might be useful so let's comment it out for now
	/*rawTx, err = commands.DecodeRawTransaction(ctx, rpcClient, commands.Transaction(funded.Hex))
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
	}*/

	switch {
	case isAssetIssuance == true:

		tx := funded.Hex
		// Those values don't matter and should work anywhere
		assetAmount := 1000.00000001
		tokenAmount := 0.00000001
		contractHash := "7F6475E61926B63C190CEBAB3470531EAB53B5481CBF960C3EE3164CA71E816B"

		issuedWithAsset, err := commands.RawIssueAssetWithAsset(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			assetAmount,
			assetAddress,
		)
		if err != nil {
			panic(err)
		}
		log.Printf("RawIssueAssetWithAsset OK, issued asset is %s\n", issuedWithAsset.Asset)

		issuedWithToken, err := commands.RawIssueAssetWithToken(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			assetAmount,
			tokenAmount,
			assetAddress,
			tokenAddress,
		)
		if err != nil {
			panic(err)
		}
		log.Printf("RawIssueAssetWithToken OK, issued asset is %s\n", issuedWithToken.Asset)

		issuedWithContract, err := commands.RawIssueAssetWithContract(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			assetAmount,
			assetAddress,
			contractHash,
		)
		if err != nil {
			panic(err)
		}
		log.Printf("RawIssueAssetWithContract OK, issued asset is %s\n", issuedWithContract.Asset)

		issuedWithTokenWithContract, err := commands.RawIssueAssetWithTokenWithContract(
			ctx,
			rpcClient,
			commands.Transaction(tx),
			assetAmount,
			tokenAmount,
			assetAddress,
			tokenAddress,
			contractHash,
		)
		if err != nil {
			panic(err)
		}

		log.Printf("RawIssueAssetWithTokenWithContract OK, issued asset is %s\n", issuedWithTokenWithContract.Asset)

		// We can choose to sign another transaction if needed
		toSign := issuedWithTokenWithContract.Hex
		blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(toSign))
		if err != nil {
			panic(err)
		}

		log.Printf("Blinded transaction OK\n")
		log.Printf("issuedWithTokenWithContract to sign:\n%+v", blinded) // this is ready for signing

		signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(blinded))
		if err != nil {
			panic(err)
		}
		if !signed.Complete {
			log.Printf("SignRawTransactionWithWallet failed\n") //this is expected if using a watch-only wallet
			return                                              // Just continue the test for now
		}
		log.Printf("issuedWithTokenWithContract signed:\n%+v", signed.Hex)

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

func RawReissuanceElements(ctx context.Context, assetID string) {
	var isAssetReissuance bool
	switch ctxValue := ctx.Value(KeyReissueAsset).(type) {
	case bool:
		isAssetReissuance = ctxValue
	}

	switch {
	case isAssetReissuance == true:

		rpcClient := elementsRpcClient("elements-testnet", 28433)
		if rpcClient == nil {
			panic("Invalid rpcClient")
		}

		issuanceInfo, err := commands.ListIssuances(ctx, rpcClient, commands.AssetID(assetID))
		if err != nil {
			panic(err)
		}
		log.Printf("issuanceInfo is %+v", issuanceInfo)
		i := 0
		for i < len(issuanceInfo) && issuanceInfo[i].Isreissuance == true {
			i++
		}
		entropy := issuanceInfo[i].Entropy
		tokenID := issuanceInfo[i].Token

		unspentInfo, err := commands.ListUnspentWithAssetWithMaxCount(ctx, rpcClient, nil, tokenID, 1)
		if err != nil {
			panic(err)
		}
		log.Printf("unspentinfo is %+v", unspentInfo)
		txID := unspentInfo[0].TxID
		vout := unspentInfo[0].Vout
		assetBlinder := unspentInfo[0].AssetBlinder
		tokenAmount := unspentInfo[0].Amount // there's no point not spending the whole UTXO here
		assetAmount := 1000.00000002

		changeAddress := "el1qqd98jldp2wm05ew4xte6l8kaaufjekrau5h698zc4wth5j7uft5cuntxyrx0yj5eapgq8lzjvkw6y7xezhuwy8sdltd68prcl"
		tokenAddress := "el1qqw5paxt5wgxxj0z4x75u4hu8x905ypmq33z7gkzpu06lpa7azhpaj7y5u8fpverafnzyuye9gcjpn8skyflhc56rrz3nzj3cu"
		assetAddress := "el1qqw5tm9jr9kl92t6ucg8cus9hr7chntenfvf634pf3jt79nftz4yteck9k7u8vf5c8jn2l5u3nac4a3vszpp47pv03huadvlp6"

		hex, err := commands.CreateRawTransaction(ctx, rpcClient, []commands.UTXOInfo{
			{TxID: txID, Vout: vout}, // this is the previous token output
		}, []commands.SpendInfo{
			{Address: tokenAddress, Amount: tokenAmount},
		}, []commands.AssetInfo{
			{Address: tokenAddress, Asset: tokenID},
		})
		if err != nil {
			panic(err)
		}
		log.Printf("CreateRawTransaction: %s\n", hex)

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

		reissued, err := commands.RawReissueAsset(
			ctx,
			rpcClient,
			commands.Transaction(funded.Hex),
			assetAmount,
			assetAddress,
			entropy,
			assetBlinder,
			0, // we put the token input at index 0 with createrawtransaction
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
			log.Printf("SignRawTransactionWithWallet failed\n") //this is expected if using a watch-only wallet
			return
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
