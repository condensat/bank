package messaging

import (
	"flag"

	"git.condensat.tech/bank"
)

type NatsOptions struct {
	bank.ServerOptions
}

func DefaultOptions() NatsOptions {
	return NatsOptions{
		ServerOptions: bank.ServerOptions{
			HostName: "nats",
			Port:     4222,
		},
	}
}

func OptionArgs(args *NatsOptions) {
	if args == nil {
		panic("Invalid args options")
	}

	OptionCmdArgs(flag.CommandLine, args)
}

func OptionCmdArgs(cmd *flag.FlagSet, args *NatsOptions) {
	if args == nil {
		panic("Invalid args options")
	}

	defaults := DefaultOptions()
	cmd.StringVar(&args.HostName, "natsHost", defaults.HostName, "Nats hostName (default 'nats')")
	cmd.IntVar(&args.Port, "natsPort", defaults.Port, "Nats port (default 4222)")
}
