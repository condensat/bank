package main

import (
	"context"
	"flag"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/monitor"
	"git.condensat.tech/bank/monitor/processus"

	"git.condensat.tech/bank/cache"
)

type MonitorApi struct {
	Port int
}

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  messaging.NatsOptions

	Monitor MonitorApi
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "MonitorApi")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)

	flag.IntVar(&args.Monitor.Port, "port", 5000, "Monitor api port (default 5000)")

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
	ctx = appcontext.WithProcessusGrabber(ctx, processus.NewGrabber(ctx, 15*time.Second))

	var monitorApi monitor.MonitorApi
	monitorApi.Run(ctx, args.Monitor.Port)
}
