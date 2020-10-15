package provider

import (
	"flag"
)

type NatsOptions struct {
	Protocol string
	HostName string
	Port     int
}

func DefaultOptions() NatsOptions {
	return NatsOptions{
		HostName: "nats",
		Port:     4222,
	}
}

func OptionArgs(args *NatsOptions) {
	if args == nil {
		panic("Invalid args options")
	}

	defaults := DefaultOptions()
	flag.StringVar(&args.HostName, "natsHost", defaults.HostName, "Nats hostName (default 'nats')")
	flag.IntVar(&args.Port, "natsPort", defaults.Port, "Nats port (default 4222)")
}
