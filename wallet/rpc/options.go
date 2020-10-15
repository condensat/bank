package rpc

import (
	"encoding/base64"
	"fmt"

	"git.condensat.tech/bank/wallet/common"
	"github.com/ybbus/jsonrpc"
)

type Options struct {
	common.ServerOptions
	User     string
	Password string

	Endpoint string
}

func (p *Options) rpcOption() jsonrpc.RPCClientOpts {
	var options jsonrpc.RPCClientOpts
	if len(p.User) > 0 {
		basic := fmt.Sprintf("%s:%s", p.User, p.Password)
		options.CustomHeaders = map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(basic)),
		}
	}
	return options
}
