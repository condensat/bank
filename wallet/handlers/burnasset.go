package handlers

import (
	"context"
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AssetBurn(ctx context.Context, request common.BurnRequest) (common.BurnResponse, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.AssetBurn")

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return common.BurnResponse{}, errors.New("Something's wrong with the chainHandler")
	}

	bankAddress, err := CryptoAddressNewDeposit(ctx, common.CryptoAddress{
		Chain:     request.Chain,
		AccountID: request.IssuerID,
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to CryptoAddressNewDeposit")
		return common.BurnResponse{}, ErrCantGetAddress
	}

	destAddress := bankAddress.PublicAddress
	if len(destAddress) == 0 {
		log.WithError(err).
			Error("destination address is empty")
		return common.BurnResponse{}, ErrCantGetAddress
	}

	bankAddress, err = CryptoAddressNewDeposit(ctx, common.CryptoAddress{
		Chain:     request.Chain,
		AccountID: request.IssuerID,
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to CryptoAddressNewDeposit")
		return common.BurnResponse{}, ErrCantGetAddress
	}

	changeAddress := bankAddress.PublicAddress
	if len(changeAddress) == 0 {
		log.WithError(err).
			Error("destination address is empty")
		return common.BurnResponse{}, ErrCantGetAddress
	}

	return chainHandler.BurnAsset(ctx, destAddress, changeAddress, request)
}

func OnAssetBurn(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnAssetBurn")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.BurnRequest
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Chain":    request.Chain,
				"IssuerID": request.IssuerID,
			})

			info, err := AssetBurn(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AssetBurn")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.BurnResponse{
				Chain:    info.Chain,
				IssuerID: info.IssuerID,
				TxID:     info.TxID,
			}, nil
		})
}
