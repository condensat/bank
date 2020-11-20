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

func AssetReissuance(ctx context.Context, request common.ReissuanceRequest) (common.ReissuanceResponse, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.AssetReissuance")

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return common.ReissuanceResponse{}, errors.New("Something's wrong with the chainHandler")
	}

	bankAddress, err := CryptoAddressNewDeposit(ctx, common.CryptoAddress{
		Chain:     request.Chain,
		AccountID: request.IssuerID,
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to CryptoAddressNewDeposit")
		return common.ReissuanceResponse{}, ErrCantGetAddress
	}

	destAddress := bankAddress.PublicAddress
	if len(destAddress) == 0 {
		log.WithError(err).
			Error("destination address is empty")
		return common.ReissuanceResponse{}, ErrCantGetAddress
	}

	bankAddress, err = CryptoAddressNewDeposit(ctx, common.CryptoAddress{
		Chain:     request.Chain,
		AccountID: request.IssuerID,
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to CryptoAddressNewDeposit")
		return common.ReissuanceResponse{}, ErrCantGetAddress
	}

	changeAddress := bankAddress.PublicAddress
	if len(changeAddress) == 0 {
		log.WithError(err).
			Error("destination address is empty")
		return common.ReissuanceResponse{}, ErrCantGetAddress
	}

	// This is the token output, but since we spend the whole output we can just complete amount next step
	request.TokenPublicAddress = destAddress

	return chainHandler.ReissueAsset(ctx, changeAddress, request)
}

func OnAssetReissuance(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnAssetReissuance")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.ReissuanceRequest
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Chain":    request.Chain,
				"IssuerID": request.IssuerID,
			})

			info, err := AssetReissuance(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AssetReissuance")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.ReissuanceResponse{
				Chain:     info.Chain,
				IssuerID:  info.IssuerID,
				TxID:      info.TxID,
				AssetVout: info.AssetVout,
				TokenVout: info.TokenVout,
			}, nil
		})
}
