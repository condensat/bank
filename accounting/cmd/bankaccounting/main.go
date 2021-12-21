package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"git.condensat.tech/bank/accounting"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/security"
)

type Accounting struct {
}

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options

	Accounting Accounting
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankAccounting")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))

	migrateDatabase(ctx)
	createDefaultFeeInfo(ctx)

	bankUser := createBankAccounts(ctx, model.UserEmail(args.App.BankUser))

	var service accounting.Accounting
	service.Run(ctx, bankUser)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(accounting.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate accounting models")
	}
}

func createDefaultFeeInfo(ctx context.Context) {
	db := appcontext.Database(ctx)

	defaultFeeInfo := []model.FeeInfo{
		// Fiat
		{Currency: "CHF", Minimum: 0.5, Rate: model.DefaultFeeRate},
		{Currency: "EUR", Minimum: 0.5, Rate: model.DefaultFeeRate},

		// Crypto
		{Currency: "BTC", Minimum: 0.00001000, Rate: model.DefaultFeeRate},
		{Currency: "LBTC", Minimum: 0.00001000, Rate: model.DefaultFeeRate},
		{Currency: "TBTC", Minimum: 0.00001000, Rate: model.DefaultFeeRate},

		// Liquid Asset with quote
		{Currency: "USDt", Minimum: 0.5, Rate: model.DefaultFeeRate},
		{Currency: "LCAD", Minimum: 0.5, Rate: model.DefaultFeeRate},
	}

	for _, feeInfo := range defaultFeeInfo {
		// Check FeeInfo validity
		if !feeInfo.IsValid() {
			logger.Logger(ctx).
				WithField("Method", "main.createDefaultFeeInfo").
				WithField("FeeInfo", feeInfo).
				Panic("Invalid default feeInfo")
			continue
		}
		// Do not update feeInfo since it could have been updated since creation
		if database.FeeInfoExists(db, feeInfo.Currency) {
			continue
		}
		// create default FeeInfo
		_, err := database.AddOrUpdateFeeInfo(db, feeInfo)
		if err != nil {
			logger.Logger(ctx).WithError(err).
				WithField("Method", "main.createDefaultFeeInfo").
				WithField("FeeInfo", "feeInfo").
				Error("AddOrUpdateFeeInfo failed")
			continue
		}
	}
}

func createBankAccounts(ctx context.Context, bankUserMail model.UserEmail) model.User {
	db := appcontext.Database(ctx)

	bankUser, err := database.FindUserByEmail(db, bankUserMail)
	if err != nil {
		logger.Logger(ctx).
			WithError(err).
			Panic("Failed to find BankUser")
	}
	if bankUser.ID == 0 {
		bankUser = model.User{
			Name:  "CondensatBank",
			Email: bankUserMail,
		}
		bankUser, err = database.FindOrCreateUser(db, bankUser)
		if err != nil {
			logger.Logger(ctx).
				WithError(err).
				WithField("UserID", bankUser.ID).
				WithField("Name", bankUser.Name).
				WithField("Email", bankUser.Email).
				Panic("Unable to FindOrCreateUser BankUser")
		}

		pubKey, privKey, err := security.CreateKeys(
			"CondensatBank",
			"Condensat PGP identity",
			string(bankUserMail),
		)
		if err != nil {
			panic(err)
		}

		// encrypt keys
		pubKey = model.PgpPublicKey(security.WriteSecret(ctx, string(pubKey)))
		privKey = model.PgpPrivateKey(security.WriteSecret(ctx, string(privKey)))

		_, err = database.AddUserPgp(db, bankUser.ID, pubKey, privKey)
		if err != nil {
			panic(err)
		}

	}

	// get public keys from database
	bankPgp, err := database.FindUserPgp(db, bankUser.ID)
	if err != nil {
		panic(err)
	}

	// decrypt public key
	bankPgp.PublicKey = model.PgpPublicKey(security.ReadSecret(ctx, string(bankPgp.PublicKey)))
	logger.Logger(ctx).
		Info("BankUser PGP publicKey")
	fmt.Fprintf(os.Stderr, "%s\n", bankPgp.PublicKey)

	bankPgp.PgpPrivateKey = model.PgpPrivateKey(security.ReadSecret(ctx, string(bankPgp.PgpPrivateKey)))

	return bankUser
}
