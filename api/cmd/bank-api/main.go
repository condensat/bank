package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/monitor"
	"git.condensat.tech/bank/networking"

	"git.condensat.tech/bank/api"
	"git.condensat.tech/bank/api/oauth"
	"git.condensat.tech/bank/networking/ratelimiter"
	"git.condensat.tech/bank/security"
)

type Api struct {
	Port              int
	CorsAllowedDomain string
	OAuth             oauth.Options
	WebAppURL         string

	SecureID string

	PeerRequestPerSecond ratelimiter.RateLimitInfo
	OpenSessionPerMinute ratelimiter.RateLimitInfo
}

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options

	Api Api
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BankApi")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)
	database.OptionArgs(&args.Database)

	flag.IntVar(&args.Api.Port, "port", 4242, "BankApi rpc port (default 4242)")
	flag.StringVar(&args.Api.CorsAllowedDomain, "corsAllowedDomain", "condensat.space", "Cors Allowed Domain (default condensat.space)")

	flag.StringVar(&args.Api.OAuth.Keys, "oauthkeys", "oauth.env", "OAuth env file for providers keys")
	flag.StringVar(&args.Api.OAuth.Domain, "oauthdomain", "condensat.space", "OAuth Domain for session cookies")
	flag.StringVar(&args.Api.WebAppURL, "webappurl", "https://app.condensat.space/", "WebApp URL")

	flag.StringVar(&args.Api.SecureID, "secureId", "secureid.json", "SecureID json file")

	args.Api.PeerRequestPerSecond = networking.DefaultPeerRequestPerSecond
	flag.IntVar(&args.Api.PeerRequestPerSecond.Rate, "peerRateLimit", 100, "Rate limit rate, per second, per peer connection (default 100)")

	args.Api.OpenSessionPerMinute = api.DefaultOpenSessionPerMinute
	flag.IntVar(&args.Api.OpenSessionPerMinute.Rate, "sessionRateLimit", 10, "Open session limit rate, per minute, per user (default 10)")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithWebAppURL(ctx, args.Api.WebAppURL)
	ctx = appcontext.WithHasherWorker(ctx, args.App.Hasher)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))
	ctx = appcontext.WithDatabase(ctx, database.New(args.Database))
	ctx = appcontext.WithProcessusGrabber(ctx, monitor.NewProcessusGrabber(ctx, 15*time.Second))
	ctx = appcontext.WithSecureID(ctx, security.SecureIDFromFile(args.Api.SecureID))

	ctx = networking.RegisterRateLimiter(ctx, args.Api.PeerRequestPerSecond)
	ctx = api.RegisterOpenSessionRateLimiter(ctx, args.Api.OpenSessionPerMinute)

	migrateDatabase(ctx)

	var api api.Api
	api.Run(ctx, args.Api.Port, corsAllowedOrigins(args.Api.CorsAllowedDomain), args.Api.OAuth)
}

func corsAllowedOrigins(corsAllowedDomain string) []string {
	// sub-domains wildcard
	return []string{fmt.Sprintf("https://%s.%s", "*", corsAllowedDomain)}
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
