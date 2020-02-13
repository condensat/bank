package main

import (
	"context"
	"flag"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/monitor/processus"

	"git.condensat.tech/bank/api"
	"git.condensat.tech/bank/api/ratelimiter"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/messaging"

	"git.condensat.tech/bank/kyc/webhook"
)

type WebHook struct {
	Port          int
	SynapsSecrets string

	PeerRequestPerSecond ratelimiter.RateLimitInfo
}

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  messaging.NatsOptions

	WebHook WebHook
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "KycWebHook")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)

	flag.IntVar(&args.WebHook.Port, "port", 4444, "KycWebHook webhook port (default 4444)")
	flag.StringVar(&args.WebHook.SynapsSecrets, "synapsSecrets", "", "Synaps hook secrets file")

	args.WebHook.PeerRequestPerSecond = api.DefaultPeerRequestPerSecond
	flag.IntVar(&args.WebHook.PeerRequestPerSecond.Rate, "peerRateLimit", 10, "Rate limit rate, per second, per peer connection (default 10)")

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

	ctx = api.RegisterRateLimiter(ctx, args.WebHook.PeerRequestPerSecond)

	synapsSecrets, err := webhook.FromFile(args.WebHook.SynapsSecrets)
	if err != nil {
		logger.Logger(ctx).
			WithError(err).
			WithField("SynapsSecrets", args.WebHook.SynapsSecrets).
			Panic("Failed to read synaps secrets from file")
	}

	var webHook webhook.WebHook
	webHook.Run(ctx, args.WebHook.Port, synapsSecrets)
}
