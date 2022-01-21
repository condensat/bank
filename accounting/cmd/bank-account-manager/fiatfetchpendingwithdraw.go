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
	FiatFetchPendingWithdraw = Command("fiatFetchPendingWithdraw")
)

type FiatFetchPendingWithdrawArg struct {
}

func fiatFetchPendingWithdrawArg(args *FiatFetchPendingWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("fiatFetchPendingWithdraw", flag.ExitOnError)

	return cmd
}

func fiatPrintPendingWithdraw(withdraws []common.FiatFetchPendingWithdraw) {
	if len(withdraws) == 0 {
		fmt.Printf("There's no pending withdraws for now\n")
	}

	for i, withdraw := range withdraws {
		fmt.Printf("\n\nWithdraw #%v: ", i)
		fmt.Printf("\nUserName: %v", withdraw.UserName)
		fmt.Printf("\nIBAN: %v", withdraw.IBAN)
		fmt.Printf("\nBIC: %v", withdraw.BIC)
		fmt.Printf("\nCurrency: %v", withdraw.Currency)
		fmt.Printf("\nAmount: %v", withdraw.Amount)
	}

	fmt.Println()

}

func fiatFetchPendingWithdraw(ctx context.Context, authInfo common.AuthInfo, args FiatFetchPendingWithdrawArg) error {
	withdraws, err := client.FiatFetchPendingWithdraw(ctx, authInfo)
	if err != nil {
		return err
	}

	fiatPrintPendingWithdraw(withdraws.PendingWithdraws)

	return nil
}
