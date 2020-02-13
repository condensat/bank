package main

import (
	"context"
	"flag"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/kyc/model"
	"git.condensat.tech/bank/kyc/worker"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/monitor/processus"

	"git.condensat.tech/bank/database"
)

type Worker struct {
}

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options

	Worker Worker
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "KycWorker")

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
	ctx = appcontext.WithProcessusGrabber(ctx, processus.NewGrabber(ctx, 15*time.Second))

	migrateDatabase(ctx)

	var worker worker.Worker
	worker.RegisterHandlers(ctx)
	worker.Run(ctx)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(model.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate kyc models")
	}
}
