package handlers

import (
	"context"
	"errors"

	"git.condensat.tech/bank/wallet/common"
)

const (
	ChainHandlerKey = "Key.ChainHandlerKey"
)

var (
	ErrInternalError = errors.New("Internal Error")
)

type ChainHandler interface {
	GetNewAddress(ctx context.Context, chain, account string) (string, error)
	ImportAddress(ctx context.Context, chain, account, address, pubkey string) error
	GetAddressInfo(ctx context.Context, chain, address string) (common.AddressInfo, error)
}

func ChainHandlerContext(ctx context.Context, chain ChainHandler) context.Context {
	return context.WithValue(ctx, ChainHandlerKey, chain)
}

func ChainHandlerFromContext(ctx context.Context) ChainHandler {
	if ctxChainHandler, ok := ctx.Value(ChainHandlerKey).(ChainHandler); ok {
		return ctxChainHandler
	}
	return nil
}
