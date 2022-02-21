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
	CryptoCancelWithdraw = Command("cryptoCancelWithdraw")
)

type CryptoCancelWithdrawArg struct {
	id      uint64
	comment string
}

func cryptoCancelWithdrawArg(args *CryptoCancelWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("cryptoCancelWithdraw", flag.ExitOnError)

	cmd.Uint64Var(&args.id, "id", 0, "id of the operation we're canceling")
	cmd.StringVar(&args.comment, "comment", "", "comment about the cancel operation")

	return cmd
}

func cryptoCancelWithdraw(ctx context.Context, authInfo common.AuthInfo, args CryptoCancelWithdrawArg) error {
	canceled, err := client.CryptoCancelWithdraw(ctx, authInfo, args.id, args.comment)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully canceled withdraw #%v:\n", canceled.WithdrawID)
	fmt.Printf("AccountID: %d\n", canceled.AccountID)
	fmt.Printf("Address: %s\n", canceled.PublicKey)
	fmt.Printf("Chain: %s\n", canceled.Chain)
	fmt.Printf("Amount: %v\n", canceled.Amount)

	fmt.Println()

	return nil
}
