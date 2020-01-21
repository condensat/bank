package messaging

import (
	"flag"
)

type NatsOptions struct {
	HostName string
	Port     int
}

func OptionArgs(args *NatsOptions) {
	if args == nil {
		panic("Invalid args options")
	}

	flag.StringVar(&args.HostName, "natsHost", "localhost", "Nats hostName (default 'localhost')")
	flag.IntVar(&args.Port, "natsPort", 4222, "Nats port (default 4222)")
}
