// simply push log entry to redis
package main

import (
	"context"
	"flag"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
)

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  messaging.NatsOptions
}

func parseArgs() Args {
	var args Args
	appcontext.OptionArgs(&args.App, "LoggerExample")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)

	flag.Parse()

	return args
}

func echoHandler(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	logger.Logger(ctx).
		WithField("Subject", subject).
		WithField("Method", "echoHandler").
		Infof("-> %s", string(message.Data))

	return message, nil
}

func natsClient(ctx context.Context) {
	messaging := appcontext.Messaging(ctx)
	messaging.SubscribeWorkers(ctx, "Example.Request", 8, echoHandler)

	log := logger.Logger(ctx)
	message := bank.NewMessage()
	message.Data = []byte("Hello, World!")

	for index := 0; index < 10; index++ {
		resp, err := messaging.Request(ctx, "Example.Request", message)
		if err != nil {
			log.
				WithError(err).
				Panicf("Request failed")
		}
		log.
			WithField("Method", "natsClient").
			Infof("<- %s", string(resp.Data))
	}
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))

	natsClient(ctx)
}
