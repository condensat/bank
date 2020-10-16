// Copyright 2020 Condensat Tech <contact@condensat.tech>. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

type Command string

const (
	CmdNewAddress = Command("new_address")
	CmdSignTx     = Command("sign_tx")
)

type RpcClient interface {
	CallFor(out interface{}, method string, params ...interface{}) error
}

func callCommand(rpcClient RpcClient, command Command, out interface{}, params ...interface{}) error {
	return rpcClient.CallFor(out, string(command), params...)
}
