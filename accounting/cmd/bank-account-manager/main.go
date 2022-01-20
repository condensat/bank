package main

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/messaging"

	log "github.com/sirupsen/logrus"

	dotenv "github.com/joho/godotenv"
)

func init() {
	_ = dotenv.Load()
}

func main() {
	ctx := context.Background()
	args := parseArgs(ctx)

	if len(args.Common.AuthInfo.OperatorAccount) > 0 && len(args.Common.AuthInfo.TOTP) == 0 {
		totp, err := readTOTP()
		if err != nil {
			panic(err)
		}
		args.Common.AuthInfo.TOTP = common.TOTP(totp)
	}

	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Common.Nats))

	Run(ctx, args)
}

func Run(ctx context.Context, args Args) {
	var err error
	switch args.Command {

	case FiatWithdraw:
		err = fiatWithdraw(ctx, args.Common.AuthInfo, args.FiatWithdraw)
	case FiatDeposit:
		err = fiatDeposit(ctx, args.Common.AuthInfo, args.FiatDeposit)
	case FiatFinalizeWithdraw:
		err = fiatFinalizeWithdraw(ctx, args.Common.AuthInfo, args.FiatFinalizeWithdraw)

	default:
		printUsage(1)
	}

	if err != nil {
		log.WithError(err).
			WithField("Command", args.Command).
			Error("Error while processing command.")
	}
}
