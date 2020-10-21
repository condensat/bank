package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"strconv"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/utils"
	"git.condensat.tech/bank/wallet/rpc"

	"git.condensat.tech/bank/wallet/bitcoin/commands"

	dotenv "github.com/joho/godotenv"
)

func init() {
	_ = dotenv.Load()
}

func main() {
	var command string
	var destAddress string
	var changeAddress string
	var assetAddress string
	var tokenAddress string
	var reissuedAsset string
	var burnAsset string
	var amountBurn float64

	flag.StringVar(&command, "command", "", "Sub command to start")

	flag.StringVar(&destAddress, "dest", "", "Address to send L-BTC")
	flag.StringVar(&changeAddress, "change", "", "Address to send change")
	flag.StringVar(&assetAddress, "asset", "", "Address to send asset")
	flag.StringVar(&tokenAddress, "token", "", "Address to send token")
	flag.StringVar(&reissuedAsset, "reissue", "", "Asset to reissue")
	flag.StringVar(&burnAsset, "burnAsset", "", "Asset to burn")
	flag.Float64Var(&amountBurn, "burnAmount", 0.0, "Amount of assets to burn")
	flag.Parse()

	ctx := context.Background()

	var err error
	switch command {

	// Bitcoin Standard

	case "RawTransactionBitcoin":
		err = RawTransactionBitcoin(ctx)

	// Liquid Elements

	case "RawTransactionElements":
		err = RawTransactionElements(ctx,
			destAddress,
			changeAddress,
			assetAddress,
			tokenAddress,
		)

	// Liquid Assets

	case "AssetIssuance":
		err = AssetIssuance(ctx,
			destAddress,
			changeAddress,
			assetAddress,
			tokenAddress,
		)

	case "Reissuance":
		err = Reissuance(ctx, reissuedAsset)

	case "BurnAsset":
		err = BurnAsset(ctx,
			destAddress,
			changeAddress,
			burnAsset,
			amountBurn,
		)

	default:
		log.Fatalf("Unknown command %s.", command)
	}

	if err != nil {
		log.Fatalf("Failed to process command. %v", err)
	}
}

// Bitcoin Standard

func RawTransactionBitcoin(ctx context.Context) error {
	rpcClient := bitcoinRpcClient()

	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: "bcrt1qjlw9gfrqk0w2ljegl7vwzrt2rk7sst8d4hm7n9", Amount: 0.003},
	}, nil)
	if err != nil {
		return err
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	rawTx, err := commands.DecodeRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		return err
	}
	decoded, err := commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		return err
	}
	log.Printf("DecodeRawTransaction: %+v\n", decoded)

	funded, err := commands.FundRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		return err
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	rawTx, err = commands.DecodeRawTransaction(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		return err
	}
	decoded, err = commands.ConvertToRawTransactionBitcoin(rawTx)
	if err != nil {
		return err
	}
	log.Printf("FundRawTransaction Hex: %+v\n", decoded)

	addressMap := make(map[commands.Address]commands.Address)
	for _, in := range decoded.Vin {

		txInfo, err := commands.GetTransaction(ctx, rpcClient, in.Txid, true)
		if err != nil {
			return err
		}

		addressMap[txInfo.Address] = txInfo.Address
		for _, d := range txInfo.Details {
			address := commands.Address(d.Address)
			addressMap[address] = address
		}
	}

	signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		return err
	}
	if !signed.Complete {
		return errors.New("SignRawTransactionWithWallet failed")
	}
	log.Printf("Signed transaction is: %+v\n", signed.Hex)

	accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
	if err != nil {
		return err
	}

	log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)
	if !accepted.Allowed {
		log.Printf("Reject-reason: %+v", accepted.Reason)
		return errors.New("TestMempoolAccept failed")
	}

	return nil
}

// Liquid Elements

func RawTransactionElements(ctx context.Context, destAddress, changeAddress, assetAddress, tokenAddress string) error {
	rpcClient := elementsRpcClient()

	// We create 2 LBTC outputs, which might be a bit unnecessary
	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: destAddress, Amount: 0.001},
	}, nil)
	if err != nil {
		return err
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	rawTx, err := commands.DecodeRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		return err
	}
	decoded, err := commands.ConvertToRawTransactionLiquid(rawTx)
	if err != nil {
		return err
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
		return err
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(funded.Hex))
	if err != nil {
		return err
	}

	log.Printf("Blinded transaction OK\n")

	signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(blinded))
	if err != nil {
		return err
	}
	if !signed.Complete {
		return errors.New("SignRawTransactionWithWallet failed")
	}

	accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
	if err != nil {
		return err
	}

	log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)
	if !accepted.Allowed {
		log.Printf("Reject-reason: %+v", accepted.Reason)
		return errors.New("TestMempoolAccept failed")
	}

	return nil
}

// Liquid Assets

func AssetIssuance(ctx context.Context, destAddress, changeAddress, assetAddress, tokenAddress string) error {
	rpcClient := elementsRpcClient()

	// We create 2 LBTC outputs, which might be a bit unnecessary
	hex, err := commands.CreateRawTransaction(ctx, rpcClient, nil, []commands.SpendInfo{
		{Address: destAddress, Amount: 0.001},
	}, nil)
	if err != nil {
		return err
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	rawTx, err := commands.DecodeRawTransaction(ctx, rpcClient, hex)
	if err != nil {
		return err
	}
	decoded, err := commands.ConvertToRawTransactionLiquid(rawTx)
	if err != nil {
		return err
	}
	log.Printf("DecodeRawTransaction: %+v\n", decoded)

	funded, err := commands.FundRawTransactionWithOptions(ctx, rpcClient,
		hex,
		commands.FundRawTransactionOptions{
			ChangeAddress:   changeAddress,
			IncludeWatching: true,
		},
	)
	if err != nil {
		return err
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	tx := funded.Hex
	// Those values don't matter and should work anywhere
	assetAmount := 1000.00000001
	tokenAmount := 0.00000001
	contractHash := "7F6475E61926B63C190CEBAB3470531EAB53B5481CBF960C3EE3164CA71E816B"

	issuedWithAsset, err := commands.RawIssueAssetWithAsset(ctx, rpcClient,
		commands.Transaction(tx),
		assetAmount,
		assetAddress,
	)
	if err != nil {
		return err
	}
	log.Printf("RawIssueAssetWithAsset OK, issued asset is %s\n", issuedWithAsset.Asset)

	issuedWithToken, err := commands.RawIssueAssetWithToken(ctx, rpcClient,
		commands.Transaction(tx),
		assetAmount,
		tokenAmount,
		assetAddress,
		tokenAddress,
	)
	if err != nil {
		return err
	}
	log.Printf("RawIssueAssetWithToken OK, issued asset is %s\n", issuedWithToken.Asset)

	issuedWithContract, err := commands.RawIssueAssetWithContract(ctx, rpcClient,
		commands.Transaction(tx),
		assetAmount,
		assetAddress,
		contractHash,
	)
	if err != nil {
		return err
	}
	log.Printf("RawIssueAssetWithContract OK, issued asset is %s\n", issuedWithContract.Asset)

	issuedWithTokenWithContract, err := commands.RawIssueAssetWithTokenWithContract(ctx, rpcClient,
		commands.Transaction(tx),
		assetAmount,
		tokenAmount,
		assetAddress,
		tokenAddress,
		contractHash,
	)
	if err != nil {
		return err
	}

	log.Printf("RawIssueAssetWithTokenWithContract OK, issued asset is %s\n", issuedWithTokenWithContract.Asset)

	// We can choose to sign another transaction if needed
	toSign := issuedWithTokenWithContract.Hex
	blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(toSign))
	if err != nil {
		return err
	}

	log.Printf("Blinded transaction OK\n")
	log.Printf("issuedWithTokenWithContract to sign:\n%+v", blinded) // this is ready for signing

	signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(blinded))
	if err != nil {
		return err
	}
	if !signed.Complete {
		return errors.New("SignRawTransactionWithWallet failed")
	}
	log.Printf("issuedWithTokenWithContract signed:\n%+v", signed.Hex)

	accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
	if err != nil {
		return err
	}

	log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)
	if !accepted.Allowed {
		log.Printf("Reject-reason: %+v", accepted.Reason)
		return errors.New("TestMempoolAccept failed")
	}

	return nil
}

func Reissuance(ctx context.Context, assetID string) error {
	rpcClient := elementsRpcClient()

	issuanceInfo, err := commands.ListIssuances(ctx, rpcClient, commands.AssetID(assetID))
	if err != nil {
		return err

	}
	log.Printf("issuanceInfo is %+v", issuanceInfo)
	if len(issuanceInfo) == 0 {
		return errors.New("Invalid OssuanceInfo")
	}

	var insuance commands.ListIssuancesInfo
	for _, info := range issuanceInfo {
		if !info.Isreissuance {
			continue
		}
		insuance = info
	}

	unspentInfo, err := commands.ListUnspentWithAssetWithMaxCount(ctx, rpcClient, nil, insuance.Token, 1)
	if err != nil {
		return err
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
		{Address: tokenAddress, Asset: insuance.Token},
	})
	if err != nil {
		return err
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	funded, err := commands.FundRawTransactionWithOptions(ctx, rpcClient,
		hex,
		commands.FundRawTransactionOptions{
			ChangeAddress:   changeAddress,
			IncludeWatching: true,
		},
	)
	if err != nil {
		return err
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	reissued, err := commands.RawReissueAsset(ctx, rpcClient,
		commands.Transaction(funded.Hex),
		assetAmount,
		assetAddress,
		insuance.Entropy,
		assetBlinder,
		0, // we put the token input at index 0 with createrawtransaction
	)
	if err != nil {
		return err
	}
	log.Printf("RawReissueAsset OK:\n%s\n", reissued)

	blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(reissued))
	if err != nil {
		return err
	}

	log.Printf("Blinded transaction OK:\n%s", blinded)

	toSign := blinded
	signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(toSign))
	if err != nil {
		return err
	}
	if !signed.Complete {
		return errors.New("SignRawTransactionWithWallet failed")
	}
	log.Printf("Reissuance TX signed:\n%+v", signed.Hex)

	accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
	if err != nil {
		return err
	}

	log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)
	if !accepted.Allowed {
		log.Printf("Reject-reason: %+v\n", accepted.Reason)
		return errors.New("TestMempoolAccept failed")
	}

	return nil
}

func BurnAsset(ctx context.Context, destAddress, changeAddress, asset string, amount float64) error {
	rpcClient := elementsRpcClient()

	log.Printf("Burning %f of asset %s\n", amount, asset)

	unspentInfo, err := commands.ListUnspentWithAsset(ctx, rpcClient, nil, asset)
	if err != nil {
		return err

	}
	var sumAmt float64
	vin := []commands.UTXOInfo{}
	for _, info := range unspentInfo {
		if sumAmt >= amount {
			break
		}
		sumAmt += info.Amount
		vin = append(vin, commands.UTXOInfo{
			TxID: info.TxID,
			Vout: info.Vout,
		})
	}
	amount = utils.ToFixed(amount, 8)
	sumAmt = utils.ToFixed(sumAmt, 8)

	if sumAmt < amount {
		return errors.New("Not enough assets to burn")
	}

	hex, err := commands.CreateRawTransaction(ctx, rpcClient, vin, []commands.SpendInfo{
		{Address: "burn", Amount: amount},
		{Address: destAddress, Amount: utils.ToFixed(sumAmt-amount, 8)},
	}, []commands.AssetInfo{
		{Address: "burn", Asset: asset},
		{Address: destAddress, Asset: asset},
	})
	if err != nil {
		return err
	}
	log.Printf("CreateRawTransaction: %s\n", hex)

	log.Printf("changeAddress is %s", changeAddress)

	funded, err := commands.FundRawTransactionWithOptions(ctx,
		rpcClient,
		hex,
		commands.FundRawTransactionOptions{
			ChangeAddress:   changeAddress,
			IncludeWatching: true,
		},
	)
	if err != nil {
		return err
	}
	log.Printf("FundRawTransaction: %+v\n", funded)

	toSign := funded.Hex
	blinded, err := commands.BlindRawTransaction(ctx, rpcClient, commands.Transaction(toSign))
	if err != nil {
		return err
	}

	log.Printf("Blinded transaction: %s\n", blinded)

	signed, err := commands.SignRawTransactionWithWallet(ctx, rpcClient, commands.Transaction(blinded))
	if err != nil {
		return err
	}
	if !signed.Complete {
		return errors.New("SignRawTransactionWithWallet failed")
	}
	log.Printf("burn transaction signed:\n%+v", signed.Hex)

	accepted, err := commands.TestMempoolAccept(ctx, rpcClient, signed.Hex)
	if err != nil {
		return err
	}

	log.Printf("Accepted in the mempool: %+v\n", accepted.Allowed)
	if !accepted.Allowed {
		log.Printf("Reject-reason: %+v", accepted.Reason)
		return errors.New("TestMempoolAccept failed")
	}

	return nil
}

// Helpers

func bitcoinRpcClient() commands.RpcClient {
	hostname := os.Getenv("BITCOIN_TESTNET_HOSTNAME")
	if len(hostname) == 0 {
		hostname = "bitcoin-testnet"
	}
	port, _ := strconv.Atoi(os.Getenv("BITCOIN_TESTNET_PORT"))
	if port == 0 {
		port = 18332
	}
	user := os.Getenv("BITCOIN_TESTNET_USER")
	if len(user) == 0 {
		user = "bank-wallet"
	}
	password := os.Getenv("BITCOIN_TESTNET_PASSWORD")

	return createRpcClient(hostname, port, user, password)
}

func elementsRpcClient() commands.RpcClient {
	hostname := os.Getenv("ELEMENTS_REGRTEST_HOSTNAME")
	if len(hostname) == 0 {
		hostname = "elements-regtest"
	}
	port, _ := strconv.Atoi(os.Getenv("ELEMENTS_REGRTEST_PORT"))
	if port == 0 {
		port = 28433
	}
	user := os.Getenv("ELEMENTS_REGRTEST_USERs")
	if len(user) == 0 {
		user = "bank-wallet"
	}
	password := os.Getenv("ELEMENTS_REGRTEST_PASSWORD")

	return createRpcClient(hostname, port, user, password)
}

func createRpcClient(hostname string, port int, user, password string) commands.RpcClient {
	rpcClient := rpc.New(rpc.Options{
		ServerOptions: bank.ServerOptions{Protocol: "http", HostName: hostname, Port: port},
		User:          user,
		Password:      password,
	}).Client

	_, err := commands.GetBlockCount(context.Background(), rpcClient)
	if err != nil {
		log.Fatalf("Rpc call failed. %s.", err)
	}

	return rpcClient
}
