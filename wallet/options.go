package wallet

import (
	"flag"
	"strings"
)

type WalletOptions struct {
	chains string
}

func (p *WalletOptions) Chains() []string {
	var result []string

	for _, chain := range strings.Split(p.chains, ",") {
		if len(chain) == 0 {
			continue
		}
		result = append(result, chain)
	}

	return result
}

func OptionArgs(args *WalletOptions) {
	if args == nil {
		panic("Invalid wallet options")
	}

	flag.StringVar(&args.chains, "chains", "bitcoin-mainnet", "Comma separated chain list (default bitcoin-mainnet,)")
}
