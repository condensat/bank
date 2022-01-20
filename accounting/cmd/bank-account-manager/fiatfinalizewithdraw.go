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
	FiatFinalizeWithdraw = Command("fiatFinalizeWithdraw")
)

type FiatFinalizeWithdrawArg struct {
	userName string
	iban     string
}

func fiatFinalizeWithdrawArg(args *FiatFinalizeWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("fiatFinalizeWithdraw", flag.ExitOnError)

	cmd.StringVar(&args.userName, "userName", "", "User that ask to withdraw money")
	cmd.StringVar(&args.iban, "iban", "", "IBAN of the recipient account")

	return cmd
}

func fiatFinalizeWithdraw(ctx context.Context, authInfo common.AuthInfo, args FiatFinalizeWithdrawArg) error {
	final, err := client.FiatFinalizeWithdraw(ctx, authInfo, args.userName, args.iban)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully finalized withdrawal from user %s to account %s\n", final.UserName, final.IBAN)

	return nil
}
