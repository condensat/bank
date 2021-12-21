package main

import (
	"context"
	"flag"

	"git.condensat.tech/bank/api"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/logger"
)

type ResultCode int

const (
	ResultCodeOK ResultCode = iota
)

type Args struct {
	App      appcontext.Options
	Database database.Options

	UserFile string
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankUserManager")
	database.OptionArgs(&args.Database)

	flag.StringVar(&args.UserFile, "userFile", "-", "UserFile or StdIn ('-')")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithHasherWorker(ctx, args.App.Hasher)
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))

	migrateDatabase(ctx)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(api.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate api models")
	}
}
