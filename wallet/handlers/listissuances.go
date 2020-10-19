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

func ListIssuances(ctx context.Context, request common.ListIssuancesRequest) ([]common.IssuanceInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.AssetIssuance")

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return []common.IssuanceInfo{}, errors.New("Something's wrong with the chainHandler")
	}

	return chainHandler.ListIssuances(ctx, request)
}

func OnListIssuances(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnListIssuances")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.ListIssuancesRequest
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Chain":    request.Chain,
				"IssuerID": request.IssuerID,
			})

			list, err := ListIssuances(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to ListIssuances")
				return nil, cache.ErrInternalError
			}

			return &common.IssuanceList{
				Chain:     request.Chain,
				IssuerID:  request.IssuerID,
				Issuances: list,
			}, nil
		})
}
