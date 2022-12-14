package bitcoin

import (
	"flag"

	"git.condensat.tech/bank"
)

type BitcoinOptions struct {
	bank.ServerOptions

	User string
	Pass string
}

func OptionArgs(args *BitcoinOptions) {
	if args == nil {
		panic("Invalid bitcoin options")
	}

	flag.StringVar(&args.HostName, "bitcoinHost", "bitcoin", "Bitcoin hostname (default 'bitcoin')")
	flag.IntVar(&args.Port, "bitcoinPort", 8332, "Bitcoin port (default 8332)")
	flag.StringVar(&args.User, "bitcoinUser", "condensat", "Bitcoin rpc user (default condensat)")
	flag.StringVar(&args.Pass, "bitcoinPass", "condensat", "Bitcoin rpc password (default condensat)")
}
