package ssm

import (
	"flag"

	"git.condensat.tech/bank"
)

type SsmOptions struct {
	bank.ServerOptions

	User string
	Pass string
}

func OptionArgs(args *SsmOptions) {
	if args == nil {
		panic("Invalid ssm options")
	}

	flag.StringVar(&args.HostName, "ssmHost", "smm", "Ssm hostname (default 'ssm')")
	flag.IntVar(&args.Port, "ssmPort", 5000, "Ssm port (default 5000)")
}
