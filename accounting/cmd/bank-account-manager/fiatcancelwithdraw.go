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
	FiatCancelWithdraw = Command("fiatCancelWithdraw")
)

type FiatCancelWithdrawArg struct {
	id      uint64
	comment string
}

func fiatCancelWithdrawArg(args *FiatCancelWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("cryptoCancelWithdraw", flag.ExitOnError)

	cmd.Uint64Var(&args.id, "id", 0, "id of the operation we're canceling")
	cmd.StringVar(&args.comment, "comment", "", "comment about the cancel operation")

	return cmd
}

func fiatCancelWithdraw(ctx context.Context, authInfo common.AuthInfo, args FiatCancelWithdrawArg) error {
	canceled, err := client.FiatCancelWithdraw(ctx, authInfo, args.id, args.comment)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully canceled withdraw #%v:\n", canceled.FiatOperationInfoID)
	fmt.Printf("UserName: %s\n", canceled.UserName)
	fmt.Printf("IBAN: %s\n", canceled.IBAN)
	fmt.Printf("Currency: %s\n", canceled.Currency)
	fmt.Printf("Amount: %v\n", canceled.Amount)

	fmt.Println()

	return nil
}
