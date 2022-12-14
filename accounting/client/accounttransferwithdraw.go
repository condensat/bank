package client

import (
	"context"
	"time"

	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountTransferWithdrawCrypto(ctx context.Context, accountID uint64, currency string, amount float64, batchMode, label string, chain, publicKey string) (uint64, error) {
	if accountID == 0 {
		return 0, cache.ErrInternalError
	}

	// Deposit amount must be positive
	if amount <= 0.0 {
		return 0, cache.ErrInternalError
	}

	if len(chain) == 0 {
		return 0, cache.ErrInternalError
	}
	if len(publicKey) == 0 {
		return 0, cache.ErrInternalError
	}

	transfer, err := accountTransferWithdrawRequest(ctx, common.AccountTransferWithdraw{
		BatchMode: batchMode,
		Source: common.AccountEntry{
			AccountID: accountID,
			Currency:  currency,

			OperationType:    "transfer",
			SynchroneousType: "sync",
			Timestamp:        time.Now(),

			Label: label,

			Amount: amount,
		},
		Crypto: common.CryptoTransfert{
			Chain:     chain,
			PublicKey: publicKey,
		},
	})
	if err != nil {
		return 0, err
	}

	return uint64(transfer.Source.ReferenceID), nil
}

func accountTransferWithdrawRequest(ctx context.Context, withdraw common.AccountTransferWithdraw) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.accountTransferWithdrawRequest")
	log = log.WithFields(logrus.Fields{
		"AccountID": withdraw.Source.AccountID,
		"Amount":    withdraw.Source.Amount,
		"Label":     withdraw.Source.Label,
	})

	var result common.AccountTransfer
	err := messaging.RequestMessage(ctx, common.AccountTransferWithdrawSubject, &withdraw, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.AccountTransfer{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SrcID":      result.Source.OperationID,
		"SrcPrevID":  result.Source.OperationPrevID,
		"SrcBalance": result.Source.Balance,

		"DstID":      result.Destination.OperationID,
		"DstPrevID":  result.Destination.OperationPrevID,
		"DstBalance": result.Destination.Balance,
	}).Debug("Withdraw request")

	return result, nil
}
