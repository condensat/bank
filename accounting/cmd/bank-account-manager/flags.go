package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/messaging"
	dotenv "github.com/joho/godotenv"
)

func init() {
	_ = dotenv.Load()
}

func printUsage(code int) {
	fmt.Println("Use command [fiatDeposit, fiatFetchPendingWithdraw, fiatFinalizeWithdraw, cryptoFetchPendingWithdraw, cryptoValidateWithdraw, cryptoCancelWithdraw, fiatCancelWithdraw]")
	os.Exit(code)
}

type Command string

type CommonArg struct {
	AuthInfo common.AuthInfo
	Nats     messaging.NatsOptions
}

type Args struct {
	Command Command
	Common  CommonArg

	FiatFetchPendingWithdraw FiatFetchPendingWithdrawArg
	FiatDeposit              FiatDepositArg
	FiatFinalizeWithdraw     FiatFinalizeWithdrawArg

	CryptoFetchPendingWithdraw CryptoFetchPendingWithdrawArg
	CryptoValidateWithdraw     CryptoValidateWithdrawArg

	CryptoCancelWithdraw CryptoCancelWithdrawArg
	FiatCancelWithdraw   FiatCancelWithdrawArg
}

func AuthInfoCmdArgs(cmd *flag.FlagSet, args *common.AuthInfo) {
	if args == nil {
		panic("Invalid args options")
	}

	cmd.StringVar(&args.OperatorAccount, "operatorAccount", "", "Operator Account")
	cmd.StringVar((*string)(&args.TOTP), "totp", "", "Operator TOTP")
}

func commonArgs(cmd *flag.FlagSet, args *CommonArg) {
	AuthInfoCmdArgs(cmd, &args.AuthInfo)
	messaging.OptionCmdArgs(cmd, &args.Nats)
}

func parseArgs(ctx context.Context) Args {
	var args Args

	if len(os.Args) == 1 {
		printUsage(1)
	}
	args.Command = Command(os.Args[1])

	var cmd *flag.FlagSet
	switch args.Command {

	case FiatDeposit:
		cmd = fiatDepositArg(&args.FiatDeposit)

	case FiatFetchPendingWithdraw:
		cmd = fiatFetchPendingWithdrawArg(&args.FiatFetchPendingWithdraw)

	case FiatFinalizeWithdraw:
		cmd = fiatFinalizeWithdrawArg(&args.FiatFinalizeWithdraw)

	case FiatCancelWithdraw:
		cmd = fiatCancelWithdrawArg(&args.FiatCancelWithdraw)

	case CryptoFetchPendingWithdraw:
		cmd = cryptoFetchPendingWithdrawArg(&args.CryptoFetchPendingWithdraw)

	case CryptoValidateWithdraw:
		cmd = cryptoValidateWithdrawArg(&args.CryptoValidateWithdraw)

	case CryptoCancelWithdraw:
		cmd = cryptoCancelWithdrawArg(&args.CryptoCancelWithdraw)

	default:
		printUsage(2)
	}

	commonArgs(cmd, &args.Common)

	// Env Overrides
	args.Common.Nats.HostName = fromStringEnv("CONDENSAT_NATS_TOR", args.Common.Nats.HostName)
	if len(args.Common.AuthInfo.OperatorAccount) == 0 {
		args.Common.AuthInfo.OperatorAccount = fromStringEnv("CONDENSAT_OPERATOR_ACCOUNT", args.Common.AuthInfo.OperatorAccount)
	}

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		printUsage(3)
	}

	return args
}

func fromStringEnv(key string, value string) string {
	e := os.Getenv(key)
	if len(e) == 0 {
		e = value
	}
	return e
}
