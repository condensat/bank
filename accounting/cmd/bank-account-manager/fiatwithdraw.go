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
	FiatWithdraw = Command("fiatWithdraw")
)

type FiatWithdrawArg struct {
	userName      string
	amount        float64
	currency      string
	withdrawLabel string
	iban          string
	bic           string
	sepaLabel     string
}

func fiatWithdrawArg(args *FiatWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("fiatWithdraw", flag.ExitOnError)

	cmd.StringVar(&args.userName, "userName", "", "User that ask to withdraw money")
	cmd.Float64Var(&args.amount, "amount", 0.0, "Amount to withdraw from the account")
	cmd.StringVar(&args.currency, "currency", "", "Currency that we intend to withdraw")
	cmd.StringVar(&args.withdrawLabel, "withdrawLabel", "", "Optional Label given by the bank")
	cmd.StringVar(&args.iban, "iban", "", "IBAN of the recipient account")
	cmd.StringVar(&args.bic, "bic", "", "BIC of the recipient account")
	cmd.StringVar(&args.sepaLabel, "sepaLabel", "", "Optional Label given by the user")

	return cmd
}

func fiatWithdraw(ctx context.Context, authInfo common.AuthInfo, args FiatWithdrawArg) error {
	operation, err := client.FiatWithdraw(ctx, authInfo, args.userName, args.amount, args.currency, args.withdrawLabel, args.iban, args.bic, args.sepaLabel)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully withdrew %.2f %s for user %s\n", operation.Amount, args.currency, args.userName)
	fmt.Printf("Destination is %s\n", args.iban)

	return nil
}
