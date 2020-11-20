package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/monitor/processus"

	"git.condensat.tech/bank/wallet/client"
	"git.condensat.tech/bank/wallet/common"

	"github.com/sirupsen/logrus"
)

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  messaging.NatsOptions
}

const (
	IssuerID uint64 = 42
)

func parseArgs() Args {
	var args Args
	appcontext.OptionArgs(&args.App, "BankWalletCli")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)

	flag.Parse()

	return args
}

func NextDeposit(ctx context.Context, chain string) error {
	log := logger.Logger(ctx).WithField("Method", "NextDeposit")

	// list all currencies
	addr, err := client.CryptoAddressNextDeposit(ctx, chain, 42)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"Chain":         addr.Chain,
		"AccountID":     addr.AccountID,
		"PublicAddress": addr.PublicAddress,
	}).Info("CryptoAddress NextDeposit")

	return nil
}

func ListIssuances(ctx context.Context, chain, asset string) error {
	log := logger.Logger(ctx).WithField("Method", "ListIssuances")

	answer, err := client.ListIssuances(ctx, chain, IssuerID, asset)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"Chain":     answer.Chain,
		"Issuer ID": answer.IssuerID,
		"Issuances": answer.Issuances,
	}).Info("ListIssuances")
	return nil
}

func AssetIssuance(ctx context.Context, chain string, issuanceMode string, assetAmount, tokenAmount float64, contractHash string) error {
	log := logger.Logger(ctx).WithField("Method", "AssetIssuance")

	var answer common.IssuanceResponse
	var issuanceError error
	switch common.AssetIssuanceMode(issuanceMode) {
	case common.AssetIssuanceModeWithAsset:
		assetAddress, err := client.CryptoAddressNewDeposit(ctx, chain, IssuerID)
		if err != nil {
			return err
		}
		answer, issuanceError = client.AssetIssuance(ctx, chain, IssuerID, assetAddress.PublicAddress, assetAmount)
		if issuanceError != nil {
			return issuanceError
		}
	case common.AssetIssuanceModeWithToken:
		assetAddress, err := client.CryptoAddressNewDeposit(ctx, chain, IssuerID)
		if err != nil {
			return issuanceError
		}
		tokenAddress, err := client.CryptoAddressNewDeposit(ctx, chain, IssuerID)
		if err != nil {
			return issuanceError
		}
		answer, issuanceError = client.AssetIssuanceWithToken(ctx, chain, IssuerID, assetAddress.PublicAddress, assetAmount, tokenAddress.PublicAddress, tokenAmount)
		if issuanceError != nil {
			return issuanceError
		}
	case common.AssetIssuanceModeWithContract:
		assetAddress, err := client.CryptoAddressNewDeposit(ctx, chain, IssuerID)
		if err != nil {
			return issuanceError
		}
		answer, issuanceError = client.AssetIssuanceWithContract(ctx, chain, IssuerID, assetAddress.PublicAddress, assetAmount, contractHash)
		if issuanceError != nil {
			return issuanceError
		}
	case common.AssetIssuanceModeWithTokenWithContract:
		assetAddress, err := client.CryptoAddressNewDeposit(ctx, chain, IssuerID)
		if err != nil {
			return issuanceError
		}
		tokenAddress, err := client.CryptoAddressNewDeposit(ctx, chain, IssuerID)
		if err != nil {
			return issuanceError
		}
		answer, issuanceError = client.AssetIssuanceWithTokenWithContract(ctx, chain, IssuerID, assetAddress.PublicAddress, assetAmount, tokenAddress.PublicAddress, tokenAmount, contractHash)
		if issuanceError != nil {
			return issuanceError
		}
	default:
		return errors.New("Unknown issuance mode")
	}

	log.WithFields(logrus.Fields{
		"Chain":     answer.Chain,
		"Issuer ID": answer.IssuerID,
		"Asset ID":  answer.AssetID,
		"Token ID":  answer.TokenID,
		"TxID":      answer.TxID,
		"Vin":       answer.Vin,
		"AssetVout": answer.AssetVout,
		"TokenVout": answer.TokenVout,
		"Entropy":   answer.Entropy,
	})
	return issuanceError

}

func AssetReissuance(ctx context.Context, chain string, assetID string, assetAmount float64) error {
	log := logger.Logger(ctx).WithField("Method", "AssetReissuance")

	var answer common.ReissuanceResponse

	if len(assetID) != 64 {
		return errors.New("Wrong asset ID, must be 64B hexstring")
	}

	if assetAmount <= 0.0 {
		return errors.New("Can't reissue null or negative asset amount")
	}

	assetAddress, err := client.CryptoAddressNewDeposit(ctx, chain, IssuerID)
	if err != nil {
		return err
	}

	answer, err = client.AssetReissuance(ctx, chain, IssuerID, assetID, assetAddress.PublicAddress, assetAmount)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"Chain":     answer.Chain,
		"Issuer ID": answer.IssuerID,
		"TxID":      answer.TxID,
		"AssetVout": answer.AssetVout,
		"TokenVout": answer.TokenVout,
	})
	return nil
}

func main() {
	var command string
	var chain string
	var contractHash string
	var issuanceMode string
	var assetAmount float64
	var tokenAmount float64
	flag.StringVar(&command, "cmd", "", "Possible commands: [getDepositAddress, listIssuances, issueAsset, reissueAsset, burnAsset]")
	flag.StringVar(&chain, "chain", "liquid-regtest", "network we'll be using")
	flag.StringVar(&issuanceMode, "issuanceMode", "", "Possible modes: [asset-only, with-token, with-contract, with-token-and-contract] (issueAsset only)")
	flag.Float64Var(&assetAmount, "assetAmount", 0.0, "amount of the new asset to issue")
	flag.Float64Var(&tokenAmount, "tokenAmount", 0.0, "amount of the reissuance token to issue(issueAsset only)")
	flag.StringVar(&contractHash, "contractHash", "", "hash to commit in the issuance(issueAsset only)")
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))
	ctx = appcontext.WithProcessusGrabber(ctx, processus.NewGrabber(ctx, 15*time.Second))

	var err error
	switch command {
	case "getDepositAddress":
		NextDeposit(ctx, chain)
	case "listIssuances":
		err = ListIssuances(ctx, chain, assetID)
	case "issueAsset":
		err = AssetIssuance(ctx, chain, issuanceMode, assetAmount, tokenAmount, contractHash)
	case "reissueAsset":
		err = AssetReissuance(ctx, chain, assetID, assetAmount)
	default:
		log.Fatalf("Unknown command %s", command)
	}
	if err != nil {
		log.Fatalf("Command %s failed with error %v", command, err)
	}
}
