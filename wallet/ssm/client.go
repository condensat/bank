package ssm

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"

	"git.condensat.tech/bank"
	"github.com/ybbus/jsonrpc"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/wallet/rpc"
	"git.condensat.tech/bank/wallet/ssm/commands"
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

func NewWithTorEndpoint(ctx context.Context, endpoint string) *SsmClient {
	proxyURL, err := url.Parse("socks5://127.0.0.1:9050")
	if err != nil {
		panic(err)
	}

	return &SsmClient{
		client: &rpc.Client{
			Client: jsonrpc.NewClientWithOpts(endpoint, &jsonrpc.RPCClientOpts{
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						Proxy: http.ProxyURL(proxyURL),
					},
				},
			}),
		},
	}
}

func (p *SsmClient) NewAddress(ctx context.Context, ssmPath commands.SsmPath) (string, error) {
	log := logger.Logger(ctx).WithField("Method", "ssm.NewAddress")

	client := p.client
	if p.client == nil {
		return "", ErrInternalError
	}

	result, err := commands.NewAddress(ctx, client.Client, ssmPath.Chain, ssmPath.Fingerprint, ssmPath.Path)
	if err != nil {
		log.WithError(err).Error("NewAddress failed")
		return "", ErrRPCError
	}

	log.
		WithField("Chain", result.Chain).
		WithField("Address", result.Address).
		Debug("SSM RPC")

	return result.Address, nil
}

func (p *SsmClient) SignTx(ctx context.Context, chain, inputransaction string, inputs ...commands.SignTxInputs) (string, error) {
	log := logger.Logger(ctx).WithField("Method", "ssm.SignTx")

	client := p.client
	if p.client == nil {
		return "", ErrInternalError
	}

	result, err := commands.SignTx(ctx, client.Client, chain, inputransaction, inputs...)
	if err != nil {
		log.WithError(err).Error("SignTx failed")
		return "", ErrRPCError
	}

	log.
		WithField("Chain", result.Chain).
		WithField("SignedTx", result.SignedTx).
		Debug("SSM RPC")

	return result.SignedTx, nil
}
