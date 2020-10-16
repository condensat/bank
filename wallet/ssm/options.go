// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ssm

import (
	"flag"

	"git.condensat.tech/bank/wallet/common"
)

type SsmOptions struct {
	common.ServerOptions

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
