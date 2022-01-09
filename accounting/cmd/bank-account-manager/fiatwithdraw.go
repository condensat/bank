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
	accountID   uint64
	referenceID uint64
	amount      float64
	currency    string
	label       string
	iban        string
	bic         string
}

func fiatWithdrawArg(args *FiatWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("fiatWithdraw", flag.ExitOnError)

	cmd.Uint64Var(&args.accountID, "accountID", 0, "Account from which we withdraw money")
	cmd.Uint64Var(&args.referenceID, "referenceID", 0, "")
	cmd.Float64Var(&args.amount, "amount", 0.0, "Amount to withdraw from the account")
	cmd.StringVar(&args.currency, "currency", "", "Currency that we intend to withdraw")
	cmd.StringVar(&args.label, "label", "", "Label of the recipient of the withdraw")
	cmd.StringVar(&args.iban, "iban", "", "IBAN of the recipient account")
	cmd.StringVar(&args.bic, "bic", "", "BIC of the recipient account")

	return cmd
}

func fiatWithdraw(ctx context.Context, authInfo common.AuthInfo, args FiatWithdrawArg) error {
	operation, err := client.FiatWithdraw(ctx, authInfo, args.accountID, args.referenceID, args.amount, args.currency, args.label, args.iban, args.bic)
	if err != nil {
		return err
	}

	fmt.Printf("Fiat Withdraw complete:\nOperationID: %d\nRecipient label:\n%s\n", operation.OperationID, args.label)

	return nil
}
