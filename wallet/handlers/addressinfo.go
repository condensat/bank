package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AddressInfo(ctx context.Context, address common.AddressInfo) (common.AddressInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.AddressInfo")
	var result common.AddressInfo

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return result, ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"Chain":         address.Chain,
		"PublicAddress": address.PublicAddress,
	})

	if len(address.Chain) == 0 {
		log.WithError(ErrInvalidChain).
			Debug("AddressNext Failed")
		return result, ErrInvalidChain
	}
	if len(address.PublicAddress) == 0 {
		log.WithError(database.ErrInvalidPublicAddress).
			Debug("AddressInfo Failed")
		return result, database.ErrInvalidPublicAddress
	}

	result, err := chainHandler.GetAddressInfo(ctx, address.Chain, address.PublicAddress)
	if err != nil {
		log.WithError(err).
			Debug("GetAddressInfo Failed")
		return result, database.ErrInvalidPublicAddress
	}
	result = common.AddressInfo{
		Chain:          address.Chain,
		PublicAddress:  address.PublicAddress,
		Unconfidential: result.Unconfidential,
		IsValid:        result.IsValid,
	}

	return result, err
}

func OnAddressInfo(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnAddressInfo")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AddressInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Chain":         request.Chain,
				"PublicAddress": request.PublicAddress,
			})

			info, err := AddressInfo(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AddressInfo")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.AddressInfo{
				Chain:          info.Chain,
				PublicAddress:  info.PublicAddress,
				Unconfidential: info.Unconfidential,
				IsValid:        info.IsValid,
			}, nil
		})
}
