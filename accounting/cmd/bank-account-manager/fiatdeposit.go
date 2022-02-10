// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"

	"git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/accounting/common"
)

const (
	FiatDeposit = Command("fiatDeposit")
)

type FiatDepositArg struct {
	userName string
	amount   float64
	currency string
	label    string
}

func fiatDepositArg(args *FiatDepositArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("fiatDeposit", flag.ExitOnError)

	cmd.StringVar(&args.userName, "userName", "", "User that deposits money")
	cmd.Float64Var(&args.amount, "amount", 0.0, "Amount to deposit on the account")
	cmd.StringVar(&args.currency, "currency", "", "Currency that we intend to deposit")
	cmd.StringVar(&args.label, "label", "", "Optional label")

	return cmd
}

func fiatDeposit(ctx context.Context, authInfo common.AuthInfo, args FiatDepositArg) error {
	operation, err := client.FiatDeposit(ctx, authInfo, args.userName, args.amount, args.currency, args.label)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully deposited %.2f %s for user %s\n", operation.Amount, operation.Currency, args.userName)

	return nil
}
