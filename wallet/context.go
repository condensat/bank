package wallet

import (
	"context"
	"sync"

	"git.condensat.tech/bank/wallet/common"
)

const (
	ChainClientKey = "Key.ChainClientKey"
)

type ChainClient interface {
	GetNewAddress(ctx context.Context, account string) (string, error)
	GetBlockCount(ctx context.Context) (int64, error)
	ListUnspent(ctx context.Context, minConf, maxConf int, addresses ...string) ([]common.AddressInfo, error)
}

func ChainClientContext(ctx context.Context, chain string, client ChainClient) context.Context {
	// check valid client
	if client == nil {
		// NOOP
		return ctx
	}
	// check if client is registered
	if client := ChainClientFromContext(ctx, chain); client != nil {
		// NOOP
		return ctx
	}

	// check if multiChainClient is presnet in context
	switch chains := ctx.Value(ChainClientKey).(type) {

	case *multiChainClient:
		// add client if not found
		if chains.Client(chain) == nil {
			chains.Add(chain, client)
		}
		return ctx

	default:
		// create pool
		ctx := context.WithValue(ctx, ChainClientKey, &multiChainClient{
			clients: make(map[string]ChainClient),
		})

		// add client to pool
		return ChainClientContext(ctx, chain, client)
	}
}

func ChainClientFromContext(ctx context.Context, chain string) ChainClient {
	switch chains := ctx.Value(ChainClientKey).(type) {
	case *multiChainClient:

		// return client form pool (can be null)
		return chains.Client(chain)

	default:
		return nil
	}
}

// Chainclient pool

type multiChainClient struct {
	sync.Mutex
	clients map[string]ChainClient
}

func (p *multiChainClient) Add(chain string, client ChainClient) {
	p.Lock()
	defer p.Unlock()

	p.clients[chain] = client
}

func (p *multiChainClient) Client(chain string) ChainClient {
	p.Lock()
	defer p.Unlock()

	client := p.clients[chain]
	return client
}
