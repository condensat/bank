package client

import (
	"context"

	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountTransferWithdrawFiat(ctx context.Context, userID, accountID uint64, currency string, amount float64, batchMode, iban, bic, label string) (uint64, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.accountTransferWithdrawFiat")
	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
		"Amount":    amount,
		"Label":     label,
	})

	if userID == 0 {
		return 0, cache.ErrInternalError
	}

	// Deposit amount must be positive
	if amount <= 0.0 {
		return 0, cache.ErrInternalError
	}

	if len(iban) == 0 {
		return 0, cache.ErrInternalError
	}
	if len(bic) == 0 {
		return 0, cache.ErrInternalError
	}

	dstIban := common.IBAN(iban)
	dstBic := common.BIC(bic)

	var result common.AccountTransfer

	withdraw := common.AccountTransferWithdrawFiat{
		BatchMode: batchMode,
		UserID:    userID,
		Source: common.AccountEntry{
			AccountID: accountID,
			Currency:  currency,

			OperationType:    "transfer",
			SynchroneousType: "sync",
			Timestamp:        common.Timestamp(),

			Label: label,

			Amount: amount,
		},
		Sepa: common.FiatSepaInfo{
			IBAN:  dstIban,
			BIC:   dstBic,
			Label: "", // We'll see how to add labels specific to the sepa beneficiary later
		},
	}

	err := messaging.RequestMessage(ctx, common.AccountTransferWithdrawFiatSubject, &withdraw, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return 0, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SrcID":      result.Source.OperationID,
		"SrcPrevID":  result.Source.OperationPrevID,
		"SrcBalance": result.Source.Balance,

		"DstID":      result.Destination.OperationID,
		"DstPrevID":  result.Destination.OperationPrevID,
		"DstBalance": result.Destination.Balance,
	}).Debug("Withdraw request")

	return uint64(result.Source.ReferenceID), nil
}

func AccountTransferWithdrawCrypto(ctx context.Context, accountID uint64, currency string, amount float64, batchMode, label, chain, publicKey string) (uint64, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.accountTransferWithdrawCrypto")
	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
		"Amount":    amount,
		"Label":     label,
	})

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

	var result common.AccountTransfer

	withdraw := common.AccountTransferWithdrawCrypto{
		BatchMode: batchMode,
		Source: common.AccountEntry{
			AccountID: accountID,
			Currency:  currency,

			OperationType:    "transfer",
			SynchroneousType: "sync",
			Timestamp:        common.Timestamp(),

			Label: label,

			Amount: amount,
		},
		Crypto: common.CryptoTransfert{
			Chain:     chain,
			PublicKey: publicKey,
		},
	}

	err := messaging.RequestMessage(ctx, common.AccountTransferWithdrawCryptoSubject, &withdraw, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return 0, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"SrcID":      result.Source.OperationID,
		"SrcPrevID":  result.Source.OperationPrevID,
		"SrcBalance": result.Source.Balance,

		"DstID":      result.Destination.OperationID,
		"DstPrevID":  result.Destination.OperationPrevID,
		"DstBalance": result.Destination.Balance,
	}).Debug("Withdraw request")

	return uint64(result.Source.ReferenceID), nil
}
