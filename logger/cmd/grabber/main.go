// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Logger grabber fetch entries from redis
package main

import (
	"context"
	"flag"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/cache"

	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/messaging/provider"
	mprovider "git.condensat.tech/bank/messaging/provider"

	"git.condensat.tech/bank/monitor"
)

type Args struct {
	App          appcontext.Options
	WithDatabase bool

	Redis    cache.RedisOptions
	Nats     mprovider.NatsOptions
	Database database.Options
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "LogGrabber")
	flag.BoolVar(&args.WithDatabase, "withDatabase", false, "Store log to database (default false)")

	mprovider.OptionArgs(&args.Nats)
	cache.OptionArgs(&args.Redis)
	database.OptionArgs(&args.Database)

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = cache.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = messaging.WithMessaging(ctx, provider.NewNats(ctx, args.Nats))
	ctx = appcontext.WithProcessusGrabber(ctx, monitor.NewProcessusGrabber(ctx, 15*time.Second))

	if args.WithDatabase {
		ctx = appcontext.WithDatabase(ctx, database.New(args.Database))
		databaseLogger := logger.NewDatabaseLogger(ctx)
		ctx = appcontext.WithLogger(ctx, databaseLogger)
		defer databaseLogger.Close()
	}

	logger := logger.NewRedisLogger(ctx)
	// Start the log grabber
	logger.Grab(ctx)
}
