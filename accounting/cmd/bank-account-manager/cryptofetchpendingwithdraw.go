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
	CryptoFetchPendingWithdraw = Command("cryptoFetchPendingWithdraw")
)

type CryptoFetchPendingWithdrawArg struct {
}

func cryptoFetchPendingWithdrawArg(args *CryptoFetchPendingWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("fiatFetchPendingWithdraw", flag.ExitOnError)

	return cmd
}
func cryptoPrintPendingWithdraw(withdraws []common.CryptoWithdraw) {
	if len(withdraws) == 0 {
		fmt.Printf("There's no pending withdraws for now\n")
	}

	for _, withdraw := range withdraws {
		fmt.Printf("\n\nWithdraw #%v: ", withdraw.TargetID) // We give the targetID to user, since it is the one he needs to validate the withdraw
		fmt.Printf("\nUserName: %v", withdraw.UserName)
		fmt.Printf("\nAddress: %v", withdraw.Address)
		fmt.Printf("\nCurrency: %v", withdraw.Currency)
		fmt.Printf("\nAmount: %v", withdraw.Amount)
	}

	fmt.Println()

}

func cryptoFetchPendingWithdraw(ctx context.Context, authInfo common.AuthInfo, args CryptoFetchPendingWithdrawArg) error {
	withdraws, err := client.CryptoFetchPendingWithdraw(ctx, authInfo)
	if err != nil {
		return err
	}

	cryptoPrintPendingWithdraw(withdraws.PendingWithdraws)

	return nil
}
