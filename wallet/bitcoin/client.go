package bitcoin

import (
	"context"
	"errors"
	"fmt"

	"git.condensat.tech/bank/logger"
	"github.com/btcsuite/btcd/chaincfg"
	rpc "github.com/btcsuite/btcd/rpcclient"
)

var (
	ErrInternalError = errors.New("Internal Error")
	ErrRPCError      = errors.New("RPC Error")
)

type BitcoinClient struct {
	conn   *rpc.ConnConfig
	client *rpc.Client
	params *chaincfg.Params
}

func paramsFromRPCPort(port int) *chaincfg.Params {
	params := &chaincfg.MainNetParams
	if port == 18332 {
		params = &chaincfg.TestNet3Params
	}
	return params
}

func New(ctx context.Context, options BitcoinOptions) *BitcoinClient {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.New")
	connCfg := &rpc.ConnConfig{
		Host:         fmt.Sprintf("%s:%d", options.HostName, options.Port),
		User:         options.User,
		Pass:         options.Pass,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	client, err := rpc.New(connCfg, nil)
	if err != nil {
		log.WithError(err).
			Error("Failed to connect to bitcoin rpc server")
	}

	return &BitcoinClient{
		conn:   connCfg,
		client: client,
		params: paramsFromRPCPort(options.Port),
	}
}

func (p *BitcoinClient) GetBlockCount(ctx context.Context) (int64, error) {
	log := logger.Logger(ctx).WithField("Method", "bitcoin.GetBlockCount")
	client := p.client
	if p.client == nil {
		return 0, ErrInternalError
	}

	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.WithError(err).Error("GetBlockCount failed")
		return blockCount, ErrRPCError
	}

	log.
		WithField("BlockCount", blockCount).
		Debug("Bitcoin RPC")

	return blockCount, nil
}
