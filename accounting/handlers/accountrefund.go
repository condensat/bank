package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"github.com/sirupsen/logrus"
)

func AccountRefund(ctx context.Context, db bank.Database, transfer common.AccountTransfer) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.accountRefund")

	log = log.WithFields(logrus.Fields{
		"SrcAccountID": transfer.Source.AccountID,
		"DstAccountID": transfer.Destination.AccountID,
		"Currency":     transfer.Source.Currency,
		"Amount":       transfer.Source.Amount,
	})

	// check operation type
	if model.OperationType(transfer.Destination.OperationType) != model.OperationTypeRefund {
		log.
			Error("OperationType is not refund")
		return common.AccountTransfer{}, database.ErrInvalidAccountOperation
	}
	// check for accounts
	if transfer.Source.AccountID == transfer.Destination.AccountID {
		log.
			Error("Can not transfer within same account")
		return common.AccountTransfer{}, database.ErrInvalidAccountOperation
	}

	// check for currencies match
	{
		// fetch source account from DB
		srcAccount, err := database.GetAccountByID(db, model.AccountID(transfer.Source.AccountID))
		if err != nil {
			log.WithError(err).
				Error("Failed to get srcAccount")
			return common.AccountTransfer{}, database.ErrInvalidAccountOperation
		}
		// fetch destination account from DB
		dstAccount, err := database.GetAccountByID(db, model.AccountID(transfer.Destination.AccountID))
		if err != nil {
			log.WithError(err).
				Error("Failed to get dstAccount")
			return common.AccountTransfer{}, database.ErrInvalidAccountOperation
		}
		// currency must match
		if srcAccount.CurrencyName != dstAccount.CurrencyName {
			log.WithFields(logrus.Fields{
				"SrcCurrency": srcAccount.CurrencyName,
				"DstCurrency": dstAccount.CurrencyName,
			}).Error("Can not transfer currencies")
			return common.AccountTransfer{}, database.ErrInvalidAccountOperation
		}
	}

	// Acquire Locks for source and destination accounts
	lockSource, err := cache.LockAccount(ctx, transfer.Source.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountTransfer{}, cache.ErrLockError
	}
	defer lockSource.Unlock()

	lockDestination, err := cache.LockAccount(ctx, transfer.Destination.AccountID)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock account")
		return common.AccountTransfer{}, cache.ErrLockError
	}
	defer lockDestination.Unlock()

	// Prepare data
	transfer.Source.OperationType = transfer.Destination.OperationType
	transfer.Source.ReferenceID = transfer.Destination.ReferenceID
	transfer.Source.Timestamp = transfer.Destination.Timestamp
	transfer.Source.Label = transfer.Destination.Label

	transfer.Source.SynchroneousType = transfer.Destination.SynchroneousType
	transfer.Source.Amount = -transfer.Destination.Amount // do not create money
	transfer.Source.LockAmount = transfer.Source.Amount   // unlock funds
	transfer.Destination.LockAmount = 0.0

	// Store operations
	var operations []model.AccountOperation
	opSrc, err := database.TxAppendAccountOperation(db,
		common.ConvertEntryToOperation(transfer.Source))
	if err != nil {
		log.WithError(err).
			Error("Failed to AppendAccountOperationSlice")
		return common.AccountTransfer{}, err
	}
	operations = append(operations, opSrc)
	opDst, err := database.TxAppendAccountOperation(db, common.ConvertEntryToOperation(transfer.Destination))
	if err != nil {
		log.WithError(err).
			Error("Failed to AppendAccountOperationSlice")
		return common.AccountTransfer{}, err
	}
	operations = append(operations, opDst)

	// response should contains 2 operations
	if len(operations) != 2 {
		log.
			Error("Invalid operations count")
		return common.AccountTransfer{}, database.ErrInvalidAccountOperation
	}

	source := operations[0]
	destination := operations[1]
	log.Trace("Account transfer")

	return common.AccountTransfer{
		Source:      common.ConvertOperationToEntry(source, "N/A"),
		Destination: common.ConvertOperationToEntry(destination, "N/A"),
	}, nil
}
