package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"git.condensat.tech/bank/appcontext"

	"git.condensat.tech/bank/backoffice"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/monitor/processus"
)

type BackOffice struct {
	Port              int
	CorsAllowedDomain string
}

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options

	BackOffice BackOffice
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BackOffice")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

	flag.IntVar(&args.BackOffice.Port, "port", 4242, "BankApi rpc port (default 4242)")
	flag.StringVar(&args.BackOffice.CorsAllowedDomain, "corsAllowedDomain", "condensat.space", "Cors Allowed Domain (default condensat.space)")

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

	var backOffice backoffice.BackOffice
	backOffice.Run(ctx, args.BackOffice.Port, corsAllowedOrigins(args.BackOffice.CorsAllowedDomain))
}

func corsAllowedOrigins(corsAllowedDomain string) []string {
	// sub-domains wildcard
	return []string{fmt.Sprintf("https://%s.%s", "*", corsAllowedDomain)}
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(backoffice.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate backoffice models")
	}
}
