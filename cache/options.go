package cache

import (
	"flag"

	"git.condensat.tech/bank"
)

type RedisOptions struct {
	bank.ServerOptions
}

func OptionArgs(args *RedisOptions) {
	if args == nil {
		panic("Invalid redis options")
	}

	flag.StringVar(&args.HostName, "redisHost", "localhost", "Redis hostName (default 'localhost')")
	flag.IntVar(&args.Port, "redisPort", 6379, "Redis port (default 6379)")
}
