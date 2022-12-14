package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func AccountTransfer(ctx context.Context, transfer common.AccountTransfer) (common.AccountTransfer, error) {
	db := appcontext.Database(ctx)

	var result common.AccountTransfer
	err := db.Transaction(func(db bank.Database) error {
		var txErr error
		result, txErr = AccountTransferWithDatabase(ctx, db, transfer)
		return txErr
	})

	return result, err
}

func AccountTransferWithDatabase(ctx context.Context, db bank.Database, transfer common.AccountTransfer) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountTransfer")

	log = log.WithFields(logrus.Fields{
		"SrcAccountID": transfer.Source.AccountID,
		"DstAccountID": transfer.Destination.AccountID,
		"Currency":     transfer.Source.Currency,
		"Amount":       transfer.Source.Amount,
	})

	// check operation type
	if !isTransfertOperation(model.OperationType(transfer.Destination.OperationType)) {
		log.
			Error("OperationType is not transfer")
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

	switch transfer.Source.SynchroneousType {
	case "sync":
		transfer.Source.Amount = -transfer.Destination.Amount // do not create money
		transfer.Source.LockAmount = 0.0
	case "async-start":
		transfer.Source.Amount = 0.0                             // funds are not gone yet
		transfer.Source.LockAmount = transfer.Destination.Amount // lock funds
	case "async-end":
		transfer.Source.Amount = -transfer.Destination.Amount     // do not create money
		transfer.Source.LockAmount = -transfer.Destination.Amount // unlock funds
	}
	switch transfer.Destination.SynchroneousType {
	case "sync":
		// NOOP
	case "async-start":
		transfer.Destination.LockAmount = transfer.Destination.Amount // lock funds
	case "async-end":
		transfer.Destination.LockAmount = -transfer.Destination.Amount // unlock funds
	}

	// Store operations
	operations, err := database.TxAppendAccountOperationSlice(db,
		common.ConvertEntryToOperation(transfer.Source),
		common.ConvertEntryToOperation(transfer.Destination),
	)
	if err != nil {
		log.WithError(err).
			Error("Failed to TxAppendAccountOperationSlice")
		return common.AccountTransfer{}, err
	}

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

func OnAccountTransfer(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountTransfer")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountTransfer
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {

			response, err := AccountTransfer(ctx, request)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"SrcAccountID": request.Source.AccountID,
						"DstAccountID": request.Destination.AccountID,
					}).Errorf("Failed to AccountTransfer")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}

func isTransfertOperation(operationType model.OperationType) bool {
	switch operationType {
	case model.OperationTypeTransfer:
		fallthrough
	case model.OperationTypeTransferFee:
		return true

	default:
		return false
	}
}
