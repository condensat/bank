package ssm

import (
	"context"
	"errors"
	"sync"

	"git.condensat.tech/bank"

	"git.condensat.tech/bank/wallet/rpc"
)

var (
	ErrInternalError    = errors.New("Internal Error")
	ErrRPCError         = errors.New("RPC Error")
	ErrInvalidAccount   = errors.New("Invalid Account")
	ErrInvalidAddress   = errors.New("Invalid Address format")
	ErrLockUnspentFails = errors.New("LockUnpent Failed")
)

type SsmClient struct {
	sync.Mutex // mutex to change params while RPC

	client *rpc.Client
}

func New(ctx context.Context, options SsmOptions) *SsmClient {
	client := rpc.New(rpc.Options{
		ServerOptions: bank.ServerOptions{Protocol: "http", HostName: options.HostName, Port: options.Port},
		User:          options.User,
		Password:      options.Pass,
	})

	return &SsmClient{
		client: client,
	}
}
