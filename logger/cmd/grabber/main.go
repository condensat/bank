// Logger grabber fetch entries from redis
package main

import (
	"context"
	"flag"

	"git.condensat.tech/bank/logger"
)

type Args struct {
	AppName  string
	LogLevel string
	Redis    logger.RedisOptions
}

func parseArgs() Args {
	var args Args
	flag.StringVar(&args.AppName, "appName", "LogGrabber", "Application Name")
	flag.StringVar(&args.LogLevel, "log", "warning", "Log level [trace, debug, info, warning, error]")

	flag.StringVar(&args.Redis.HostName, "redisHost", "localhost", "Redis hostName (default 'localhost')")
	flag.IntVar(&args.Redis.Port, "redisPort", 6379, "Redis port (default 6379)")

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := logger.WithAppName(context.Background(), args.AppName)
	ctx = logger.WithLogLevel(ctx, args.LogLevel)

	redisLogger := logger.NewRedisLogger(args.Redis)
	// Start the log grabber
	redisLogger.Grab(ctx)
}
