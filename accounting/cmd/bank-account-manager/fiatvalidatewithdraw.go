// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/accounting/common"
)

const (
	FiatValidateWithdraw = Command("fiatValidateWithdraw")
)

type FiatValidateWithdrawArg struct {
	id []uint64
}

func fiatValidateWithdrawArg(args *FiatValidateWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("fiatValidateWithdraw", flag.ExitOnError)

	if len(os.Args) > 2 {
		for _, id := range os.Args[2:] {

			if id == "--help" {
				break
			}
			intID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				fmt.Printf("Provided argument \"%s\" is not parsable as int\nSkipping to next argument\n", id)
				continue
			}

			if intID <= 0 {
				fmt.Printf("Provided id %d is a negative number.\nSkipping\n", intID)
				continue
			}
			args.id = append(args.id, uint64(intID))
		}
	}

	return cmd
}

func fiatValidateWithdraw(ctx context.Context, authInfo common.AuthInfo, args FiatValidateWithdrawArg) error {
	validated, err := client.FiatValidateWithdraw(ctx, authInfo, args.id)
	if err != nil {
		return err
	}

	if len(validated.ValidatedWithdraws) > 0 {
		for _, withdraw := range validated.ValidatedWithdraws {
			fmt.Printf("Successfully validated withdraw #%v:\n", withdraw.TargetID)
			fmt.Printf("UserName: %s\n", withdraw.UserName)
			fmt.Printf("IBAN: %s\n", withdraw.IBAN)
			fmt.Printf("Amount: %v\n", withdraw.Amount)
			fmt.Printf("Currency: %s\n", withdraw.Currency)

			fmt.Println()
		}
	} else {
		fmt.Println("No valid withdraws in provided id")
	}

	return nil
}
