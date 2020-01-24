package messaging

import (
	"flag"

	"git.condensat.tech/bank"
)

type NatsOptions struct {
	bank.ServerOptions
}

func OptionArgs(args *NatsOptions) {
	if args == nil {
		panic("Invalid args options")
	}

	flag.StringVar(&args.HostName, "natsHost", "localhost", "Nats hostName (default 'localhost')")
	flag.IntVar(&args.Port, "natsPort", 4222, "Nats port (default 4222)")
}
