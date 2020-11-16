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

var (
	ErrChainClientNotFound = errors.New("Client not found")
	ErrInvalidIssuanceInfo = errors.New("Provided Issuance Info are invalid")
	ErrCantGetAddress      = errors.New("Can't get a new address")
	ErrCreatingTransaction = errors.New("Can't create transaction for issuance")
)

const (
	DefaultAmountForIssuance = 0.1
)

func AssetIssuance(ctx context.Context, request common.IssuanceRequest) (common.IssuanceResponse, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.AssetIssuance")

	// sanity check for different mode
	if !request.IsValid() {
		log.Errorf("Mode %s is invalid", string(request.Mode))
		return common.IssuanceResponse{}, ErrInvalidIssuanceInfo
	}

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return common.IssuanceResponse{}, errors.New("Something's wrong with the chainHandler")
	}

	bankAddress, err := CryptoAddressNewDeposit(ctx, common.CryptoAddress{
		Chain:     request.Chain,
		AccountID: request.IssuerID,
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to CryptoAddressNewDeposit")
		return common.IssuanceResponse{}, ErrCantGetAddress
	}

	destAddress := bankAddress.PublicAddress
	if len(destAddress) == 0 {
		log.WithError(err).
			Error("destination address is empty")
		return common.IssuanceResponse{}, ErrCantGetAddress
	}

	bankAddress, err = CryptoAddressNewDeposit(ctx, common.CryptoAddress{
		Chain:     request.Chain,
		AccountID: request.IssuerID,
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to CryptoAddressNewDeposit")
		return common.IssuanceResponse{}, ErrCantGetAddress
	}

	changeAddress := bankAddress.PublicAddress
	if len(changeAddress) == 0 {
		log.WithError(err).
			Error("destination address is empty")
		return common.IssuanceResponse{}, ErrCantGetAddress
	}

	// The amount of the LBTC output doesn't matter much, as long as it is enough to pay fees and not leave dust
	// Maybe we could first use a relatively high amount to be safe, and see later
	output := common.SpendInfo{
		PublicAddress: destAddress,
		Amount:        DefaultAmountForIssuance,
	}

	return chainHandler.IssueNewAsset(ctx, changeAddress, output, request)
}

func OnAssetIssuance(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnAssetIssuance")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.IssuanceRequest
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Chain":    request.Chain,
				"IssuerID": request.IssuerID,
			})

			info, err := AssetIssuance(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to AssetIssuance")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &common.IssuanceResponse{
				Chain:     info.Chain,
				IssuerID:  info.IssuerID,
				AssetID:   info.AssetID,
				TokenID:   info.TokenID,
				TxID:      info.TxID,
				Vin:       info.Vin,
				AssetVout: info.AssetVout,
				TokenVout: info.TokenVout,
				Entropy:   info.Entropy,
			}, nil
		})
}
