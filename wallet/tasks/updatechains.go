package tasks

import (
	"context"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/cache"
	"git.condensat.tech/bank/wallet/chain"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// UpdateChains
func UpdateChains(ctx context.Context, epoch time.Time, chains []string) {
	log := logger.Logger(ctx).WithField("Method", "task.ChainUpdate")

	chainsStates, err := chain.FetchChainsState(ctx, chains...)
	if err != nil {
		log.WithError(err).
			Error("Failed to FetchChainsState")
		return
	}

	log.WithFields(logrus.Fields{
		"Epoch": epoch.Truncate(time.Millisecond),
		"Count": len(chainsStates),
	}).Info("Chain state fetched")

	err = cache.UpdateRedisChain(ctx, chainsStates...)
	if err != nil {
		log.WithError(err).
			Error("Failed to UpdateRedisChain")
		return
	}

	for _, state := range chainsStates {
		updateChain(ctx, epoch, state)
	}
}

func updateChain(ctx context.Context, epoch time.Time, state chain.ChainState) {
	log := logger.Logger(ctx).WithField("Method", "task.ChainUpdate")
	db := appcontext.Database(ctx)

	list, addresses := fetchActiveAddresses(ctx, state)

	// Resquest chain
	infos, err := chain.FetchChainAddressesInfo(ctx, state, AddressInfoMinConfirmation, AddressInfoMaxConfirmation, list...)
	if err != nil {
		log.WithError(err).
			Error("Failed to FetchChainAddressesInfo")
		return
	}

	// local map for lookup cryptoAddresses from PublicAddress
	type CryptoTransaction struct {
		CryptoAddress model.CryptoAddress
		Transactions  []chain.TransactionInfo
	}
	cryptoTransactions := make(map[string]CryptoTransaction)

	// update firstBlockId for NextDeposit
	for _, info := range infos {
		for _, cryptoAddress := range addresses {
			// search for matching public address
			publicAddress := string(cryptoAddress.PublicAddress)
			if !matchPublicAddress(cryptoAddress, info.PublicAddress) {
				continue
			}

			// store into local map
			cryptoTransaction := CryptoTransaction{
				CryptoAddress: cryptoAddress,
				Transactions:  info.Transactions[:],
			}
			cryptoTransactions[publicAddress] = cryptoTransaction

			// update FirstBlockId
			firstBlockId := model.MemPoolBlockID // if returned FetchChainAddressesInfo, a tx exists at least in the mempool
			if info.Mined > 0 {
				firstBlockId = model.BlockID(info.Mined)
			}
			// skip if not changed
			if firstBlockId == cryptoAddress.FirstBlockId {
				continue
			}

			// update FirstBlockId
			cryptoTransaction.CryptoAddress.FirstBlockId = firstBlockId

			// store into db
			cryptoAddressUpdate, err := database.AddOrUpdateCryptoAddress(db, cryptoTransaction.CryptoAddress)
			if err != nil {
				log.WithError(err).
					Error("Failed to AddOrUpdateCryptoAddress")
			}

			// update cryptoAddress
			cryptoTransaction.CryptoAddress = cryptoAddressUpdate
			// update local map
			cryptoTransactions[publicAddress] = cryptoTransaction
			break
		}
	}

	// updateOperation transactions
	for _, cryptoTransaction := range cryptoTransactions {
		for _, transactions := range cryptoTransaction.Transactions {
			err := updateOperation(ctx, cryptoTransaction.CryptoAddress.ID, transactions)
			if err != nil {
				log.WithError(err).
					Error("Failed to updateOperation")
				continue
			}
		}
	}
}

func matchPublicAddress(crytoAddress model.CryptoAddress, address string) bool {
	if len(address) == 0 {
		return false
	}
	return string(crytoAddress.PublicAddress) == address || string(crytoAddress.Unconfidential) == address
}

func updateOperation(ctx context.Context, cryptoAddressID model.CryptoAddressID, transaction chain.TransactionInfo) error {
	log := logger.Logger(ctx).WithField("Method", "Wallet.updateOperation")
	db := appcontext.Database(ctx)

	txID := model.TxID(transaction.TxID)

	log = log.WithFields(logrus.Fields{
		"CryptoAddressID": cryptoAddressID,
		"TxID":            txID,
	})

	// create OperationInfo and update OperationStatus
	err := db.Transaction(func(db bank.Database) error {
		operationInfo, err := database.GetOperationInfoByTxId(db, txID)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.WithError(err).
				Error("Failed to GetOperationInfoByTxId")
			return err
		}

		// operationInfo does not exists
		if operationInfo.ID == 0 {
			// create new OperationInfo
			info, err := database.AddOperationInfo(db, model.OperationInfo{
				CryptoAddressID: cryptoAddressID,
				TxID:            txID,
				Amount:          model.Float(transaction.Amount),
			})
			if err != nil {
				log.WithError(err).
					Error("Failed to AddOperationInfo")
				return err
			}

			// store result
			operationInfo = info
			log.WithField("OperationID", operationInfo.ID).
				Debug("OperationInfo created")
		}

		if operationInfo.ID == 0 {
			log.
				Error("Invalid operation ID")
			return database.ErrDatabaseError
		}

		log := log.WithField("operationInfoID", operationInfo.ID)

		// create or update OperationStatus
		operationState := "received"
		if transaction.Confirmations >= ConfirmedBlockCount {
			operationState = "confirmed"
		}

		// fetch OperationStatus if exists
		status, _ := database.GetOperationStatus(db, operationInfo.ID)
		if status.Accounted == "settled" {
			operationState = status.Accounted
		}

		// check if update is needed
		if status.State == operationState {
			return nil
		}

		// update state
		status, err = database.AddOrUpdateOperationStatus(db, model.OperationStatus{
			OperationInfoID: operationInfo.ID,
			State:           operationState,
			Accounted:       status.Accounted,
		})
		if err != nil {
			log.WithError(err).
				Error("Failed to AddOrUpdateOperationStatus")
			return err
		}

		log.WithField("OperationStatus", status.State).
			Debug("OperationStatus updated")

		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to perform database transaction")
		return err
	}

	return nil
}

type AddressMap map[string]model.CryptoAddress

func addNewAddress(allAddresses AddressMap, addresses ...model.CryptoAddress) {
	for _, address := range addresses {
		publicAddress := string(address.PublicAddress)
		if _, ok := allAddresses[publicAddress]; !ok {
			allAddresses[publicAddress] = address
		}
	}
}

func fetchActiveAddresses(ctx context.Context, state chain.ChainState) ([]string, []model.CryptoAddress) {
	log := logger.Logger(ctx).WithField("Method", "task.fetchActiveAddresses")
	db := appcontext.Database(ctx)
	chainName := model.String(state.Chain)

	log = log.WithFields(logrus.Fields{
		"Chain":  state.Chain,
		"Height": state.Height,
	})

	// localMap for all unque addresses
	allAddresses := make(AddressMap)

	// fetch unused addresses from database
	{
		unused, err := database.AllUnusedCryptoAddresses(db, chainName)
		if err != nil {
			log.WithError(err).
				Error("Failed to AllUnusedCryptoAddresses")
			return nil, nil
		}

		addNewAddress(allAddresses, unused...)
	}

	// fetch mempool addresses from database
	{
		mempool, err := database.AllMempoolCryptoAddresses(db, chainName)
		if err != nil {
			log.WithError(err).
				Error("Failed to AllMempoolCryptoAddresses")
			return nil, nil
		}

		addNewAddress(allAddresses, mempool...)
	}

	// fetch unconfirmed addresses from database
	unconfirmed, err := database.AllUnconfirmedCryptoAddresses(db, chainName, model.BlockID(state.Height-UnconfirmedBlockCount))
	{
		if err != nil {
			log.WithError(err).
				Error("Failed to AllUnconfirmedCryptoAddresses")
			return nil, nil
		}

		addNewAddress(allAddresses, unconfirmed...)
	}

	// fetch missing addresses from database
	{
		missing, err := database.FindCryptoAddressesNotInOperationInfo(db, chainName)
		if err != nil {
			log.WithError(err).
				Error("Failed to FindCryptoAddressesNotInOperationInfo")
			return nil, nil
		}

		addNewAddress(allAddresses, missing...)
	}

	// fetch addresses with status received from database
	{
		received, err := database.FindCryptoAddressesByOperationInfoState(db, chainName, model.String("received"))
		if err != nil {
			log.WithError(err).
				Error("Failed to FindCryptoAddressesByOperationInfoState")
			return nil, nil
		}

		addNewAddress(allAddresses, received...)
	}

	// create final addresses lists
	var result []string                 // addresses for rpc call
	var addresses []model.CryptoAddress // addresses for operations update
	for _, cryptoAddress := range allAddresses {
		address := string(cryptoAddress.PublicAddress)
		if len(cryptoAddress.Unconfidential) != 0 {
			// use unconfidential address for listunspent call
			address = string(cryptoAddress.Unconfidential)
		}
		result = append(result, address)
		addresses = append(addresses, cryptoAddress)
	}
	return result, addresses
}
