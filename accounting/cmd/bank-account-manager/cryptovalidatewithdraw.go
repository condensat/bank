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
	CryptoValidateWithdraw = Command("cryptoValidateWithdraw")
)

type CryptoValidateWithdrawArg struct {
	id []uint64
}

func cryptoValidateWithdrawArg(args *CryptoValidateWithdrawArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("cryptoValidateWithdraw", flag.ExitOnError)

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

func cryptoValidateWithdraw(ctx context.Context, authInfo common.AuthInfo, args CryptoValidateWithdrawArg) error {
	Validated, err := client.CryptoValidateWithdraw(ctx, authInfo, args.id)
	if err != nil {
		return err
	}

	if len(Validated.ValidatedWithdraws) > 0 {
		for _, withdraw := range Validated.ValidatedWithdraws {
			fmt.Printf("Successfully validated withdraw #%v:\n", withdraw.TargetID)
			fmt.Printf("UserName: %s\n", withdraw.UserName)
			fmt.Printf("Address: %s\n", withdraw.Address)
			fmt.Printf("Currency: %s\n", withdraw.Currency)
			fmt.Printf("Amount: %v\n", withdraw.Amount)

			fmt.Println()
		}
	} else {
		fmt.Println("No valid withdraws in provided id")
	}

	return nil
}
