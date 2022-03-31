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
	"git.condensat.tech/bank/database/model"
)

const (
	CancelWithdraw = Command("cancelWithdraw")
)

type CancelWithdrawArg struct {
	id      uint64
	comment string
}

func cancelWithdrawArg(args *CancelWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("cancelWithdraw", flag.ExitOnError)

	cmd.Uint64Var(&args.id, "id", 0, "id of the operation we're canceling")
	cmd.StringVar(&args.comment, "comment", "", "comment about the cancel operation")

	return cmd
}

func cancelWithdraw(ctx context.Context, authInfo common.AuthInfo, args CancelWithdrawArg) error {
	canceled, err := client.CancelWithdraw(ctx, authInfo, args.id, args.comment)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully canceled withdraw #%v:\n", canceled.WithdrawID)
	fmt.Printf("AccountID: %d\n", canceled.AccountID)
	fmt.Printf("Amount: %v\n", canceled.Amount)
	fmt.Printf("Type: %v\n", canceled.Type)

	switch canceled.Type {
	case string(model.WithdrawTargetOnChain):
		fmt.Printf("Address: %s\n", canceled.PublicKey)
		fmt.Printf("Chain: %s\n", canceled.Chain)
	case string(model.WithdrawTargetSepa):
		fmt.Printf("IBAN: %s\n", canceled.IBAN)
	}

	fmt.Println()

	return nil
}
