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
	KeyIssueAsset   = "Key.IssueAsset"
	KeyReissueAsset = "Key.ReissueAsset"
)

func init() {
	_ = dotenv.Load()
}

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, KeyIssueAsset, true)
	ctx = context.WithValue(ctx, KeyReissueAsset, true)
	RawTransaction(ctx)
	RawTransactionElements(ctx)
	RawReissuanceElements(ctx)
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

func RawTransactionElements(ctx context.Context) {
	var isAssetIssuance bool
	switch ctxValue := ctx.Value(KeyIssueAsset).(type) {
	case bool:
		isAssetIssuance = ctxValue
	}
	rpcClient := elementsRpcClient("elements-testnet", 28433) // this may change
	if rpcClient == nil {
		panic("Invalid rpcClient")
	}

	// Below are some values that work with the wallet I use for testing, but obviously it would fail with any other
	destinationAddress := "el1qqgtvpnmmxdpp9ramzde76496m4hsvu7vtxqnhps4qp37k90r7vnvcvyajp5fef556em57lku9qju832r9ddtv8s0u2q8srlu0"
	changeAddress := "el1qq2tt6t2r5z5p99fj4z6zevs8wahdnvxgs3fn0nu2fy4ngkkdqhfpm9sgxxx76t0cphgu95s6rhjxjzvka3fxxdv4vqx9x86hz"
	assetAddress := "el1qq2l5kfqg2l9qy0nptnpz7w6rpjx4xq2njuepejgffl4jt2vd74njl5zh3rdgs3tra2z94aj69ws77aj9ar7xkuusj4pnjru7a"
	tokenAddress := "el1qq054gxp5n06fu992pavzalzmuy3xrev2299yvz68u3szf40whyy2sruncl23d5j6cswmeqt7xdetmryhztl05wl36h8pw5yxg"

	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: string(destinationAddress), Amount: 0.001},
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

func RawReissuanceElements(ctx context.Context) {
	var isAssetReissuance bool
	switch ctxValue := ctx.Value(KeyReissueAsset).(type) {
	case bool:
		isAssetReissuance = ctxValue
	}

	switch {
	case isAssetReissuance == true:

		// Those values are highly dependent on the wallet, and will fail with any other wallet
		// next step would be to get thos values from the node and not hard-coded so that it can be portable
		entropy := "fd300cee9557a6b1fb3b20d2349f68f911e4b8941ba85a098694058b68a7e0e4"
		tokenID := "91da6e2b69f5cb7a69a1c8599fd112b39be3775a9c7c5c15c7c455ec20b520da"
		txID := "41b396cf4c6d19ac17e42f18fd6f09572804e230198fc70f87da3e76ea316307"
		vout := 1
		assetBlinding := "c85be42b4b8d6afc4ffceabf8826b9e733f843d3f761daf0f15b3c6c6aa7ac89"
		// except those 2 which don't matter
		tokenAmount := 0.00000001
		assetAmount := 1000.00000002

		rpcClient := elementsRpcClient("elements-testnet", 28433)
		if rpcClient == nil {
			panic("Invalid rpcClient")
		}

		// I should be able to get the assetblinder like this:
		// call listunspent for the token and look at the asset_blinder field
		/*tokenUtxo, err := commands.ListUnspentWithAsset(ctx, rpcClient, nil, tokenID)
		if err != nil {
			panic(err)
		}
		assetBlinding := tokenUtxo[0].AssetBlinder*/

		// Those don't matter much, better to use our address though
		changeAddress := "el1qqfrxvlt6hmqjnewjtnjn3d6t0w3xgqd4h09g0r6ng4lskdrgafmzdfccykjuftdu5hcw8m56gv3g978nrcdek8nasqquddhuf"
		tokenAddress := "el1qqgxj54w6f0amzuctkptu3tgt5pxacjfjvas70tyq7m58lz7sp0tccp9cr2vxrx7v83y0r63f24lh25vxupfyc4nw4df8d6g8c"
		assetAddress := "el1qqdjjeqzweh26nx9jvztmmzzkhx3u44j4n9dfhtf89dwn79xzd2dh87m40kkr7ucfr34n544t4q39r6glh2rxzfgqnfv6jscec"

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
			assetBlinding,
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
