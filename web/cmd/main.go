package main

import (
	"context"
	"flag"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/web"

	"git.condensat.tech/bank/api"
	"git.condensat.tech/bank/api/ratelimiter"
)

type WebApp struct {
	Port      int
	Directory string

	PeerRequestPerSecond ratelimiter.RateLimitInfo
	OpenSessionPerMinute ratelimiter.RateLimitInfo
}

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions

	WebApp WebApp
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankWebApp")

	cache.OptionArgs(&args.Redis)

	flag.IntVar(&args.WebApp.Port, "port", 4420, "BankWebApp http port (default 4420)")
	flag.StringVar(&args.WebApp.Directory, "webDirectory", "/var/www", "BankWebApp http web directory (default /var/www)")

	args.WebApp.PeerRequestPerSecond = api.DefaultPeerRequestPerSecond
	flag.IntVar(&args.WebApp.PeerRequestPerSecond.Rate, "peerRateLimit", 20, "Rate limit rate, per second, per peer connection (default 20)")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))

	ctx = api.RegisterRateLimiter(ctx, args.WebApp.PeerRequestPerSecond)

	var web web.Web
	web.Run(ctx, args.WebApp.Port, args.WebApp.Directory)
}
