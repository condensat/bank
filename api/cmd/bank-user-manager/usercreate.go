// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"git.condensat.tech/bank/api/client"
	"git.condensat.tech/bank/api/common"
)

const (
	UserCreate = Command("userCreate")
)

type UserCreateArg struct {
	PGPPublicKey string
}

func userCreateArg(args *UserCreateArg) *flag.FlagSet {
	cmd := flag.NewFlagSet("userCreate", flag.ExitOnError)

	cmd.StringVar(&args.PGPPublicKey, "pgpPublicKey", "", "Client PGP public key filename")

	return cmd
}

func userCreate(ctx context.Context, authInfo common.AuthInfo, args UserCreateArg) error {
	data, err := ioutil.ReadFile(args.PGPPublicKey)
	if err != nil {
		return err
	}

	userInfo, err := client.UserCreate(ctx, authInfo,
		common.PGPPublicKey(data),
	)
	if err != nil {
		return err
	}

	fmt.Printf("User Account created:\nAccountNumber: %s\n%v\nMessage:\n%s\n", userInfo.AccountNumber, userInfo.Timestamp, userInfo.PayLoad)

	return nil
}
