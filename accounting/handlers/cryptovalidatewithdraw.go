package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/security/utils"
	"github.com/sirupsen/logrus"
)

func CryptoValidateWithdraw(ctx context.Context, authInfo common.AuthInfo, id []uint64) (common.CryptoValidatedWithdrawList, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.CryptoValidateWithdraw")
	var result common.CryptoValidatedWithdrawList

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	if common.WithOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return result, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return result, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return result, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return result, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return result, errors.New("CheckTOTP failed")
		}
		if !valid {
			return result, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return result, errors.New("Wrong operator ID")
		}
	}

	// Let's check the status of each withdraw
	var toUpdate []model.WithdrawTarget
	for _, tid := range id {
		// Look up the target
		wt, err := database.GetWithdrawTarget(db, model.WithdrawTargetID(tid))
		if err != nil {
			log.WithField("Error", err).Infoln("Can't find target in db, skipping")
			continue
		}

		// look up info by withdraw id. Is the withdraw in "created" status?
		winfo, err := database.GetLastWithdrawInfo(db, wt.WithdrawID)
		if err != nil {
			log.WithFields(logrus.Fields{
				"Error":      err,
				"WithdrawID": wt.WithdrawID,
			}).Infoln("Can't find info in db, skipping")
			continue
		}

		if winfo.Status != model.WithdrawStatusCreated {
			log.WithFields(logrus.Fields{
				"TargetID":   wt.ID,
				"WithdrawID": wt.WithdrawID,
				"Status":     winfo.Status,
			}).Infoln("Withdraw status is not in created status, skipping.")
			continue
		}

		// now we can add it to our list
		toUpdate = append(toUpdate, wt)

		// update our result list
		w, err := database.GetWithdraw(db, winfo.WithdrawID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetWithdraw")
			return result, err
		}

		// get data
		data, err := wt.OnChainData()
		if err != nil {
			log.WithError(err).
				Error("Failed to get OnChainData")
			return result, errors.New("error")
		}

		// Get userName
		accountID := w.From

		accountInfo, err := database.GetAccountByID(db, accountID)
		if err != nil {
			return result, err
		}

		userInfo, err := database.FindUserById(db, accountInfo.UserID)

		userName := userInfo.Name

		withdraw := common.CryptoWithdraw{
			WithdrawID: uint64(wt.WithdrawID),
			TargetID:   uint64(wt.ID),
			UserName:   string(userName),
			Address:    data.PublicKey,
			Currency:   string(accountInfo.CurrencyName),
			Amount:     float64(*w.Amount),
		}
		log.WithFields(logrus.Fields{
			"WithdrawID": withdraw.WithdrawID,
			"Currency":   withdraw.Currency,
			"Amount":     withdraw.Amount,
			"UserName":   withdraw.UserName,
			"Address":    withdraw.Address,
		}).Debug("Processing withdraw")

		result.ValidatedWithdraws = append(result.ValidatedWithdraws, withdraw)
	}

	// Now we can process all the withdraws
	if len(toUpdate) > 0 {
		err := ProcessWithdraws(ctx, toUpdate)
		if err != nil {
			return result, err
		}
	} else {
		log.Info("No valid withdraw to process")
	}

	return result, nil
}

type withdrawOnChainData struct {
	Withdraw model.Withdraw
	History  []model.WithdrawInfo
	Data     model.WithdrawTargetOnChainData
}

var (
	ErrProcessingWithdraw     = errors.New("Error Processing Withdraw")
	ErrProcessingWithdrawType = errors.New("Error Processing Withdraw Type")
)

func ProcessWithdraws(ctx context.Context, withdraws []model.WithdrawTarget) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.ProcessWithdraws")

	byType := make(map[model.WithdrawTargetType][]model.WithdrawTarget)

	for _, withdraw := range withdraws {
		if _, ok := byType[withdraw.Type]; !ok {
			byType[withdraw.Type] = make([]model.WithdrawTarget, 0)
		}
		byType[withdraw.Type] = append(byType[withdraw.Type], withdraw)
	}

	for _, withdraws := range byType {
		err := processWithdraws(ctx, withdraws)
		if err != nil {
			log.WithError(err).Error("Fail to processWithdraws")
		}
	}

	return nil
}

func processWithdraws(ctx context.Context, withdraws []model.WithdrawTarget) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processWithdraws")
	db := appcontext.Database(ctx)

	if len(withdraws) == 0 {
		return nil
	}

	var datas []withdrawOnChainData
	wType := withdraws[0].Type

	switch wType {
	case model.WithdrawTargetOnChain:

		// fetch withdraw info from database
		for _, withdraw := range withdraws {
			// each withdraw should have same type
			if withdraw.Type != wType {
				log.WithFields(logrus.Fields{
					"RefType":      wType,
					"WithdrawType": withdraw.Type,
				}).Error("Wrong withdraw type")
				return ErrProcessingWithdrawType
			}

			// get withdraw
			w, err := database.GetWithdraw(db, withdraw.WithdrawID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetWithdraw")
				return err
			}
			// Get withdraw info history
			history, err := database.GetWithdrawHistory(db, withdraw.WithdrawID)
			if err != nil {
				log.WithError(err).
					Error("Failed to GetWithdrawHistory")
				return ErrProcessingWithdraw
			}
			// skip processed withdraw
			if len(history) != 1 || history[0].Status != model.WithdrawStatusCreated {
				log.Warn("Withdraw status is not created")
				continue
			}

			// get data
			data, err := withdraw.OnChainData()
			if err != nil {
				log.WithError(err).
					Error("Failed to get OnChainData")
				return ErrProcessingWithdraw
			}

			datas = append(datas, withdrawOnChainData{
				Withdraw: w,
				History:  history,
				Data:     data,
			})
		}

		return processWithdrawOnChain(ctx, datas)

	default:
		return ErrProcessingWithdrawType
	}
}

func processWithdrawOnChain(ctx context.Context, datas []withdrawOnChainData) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processWithdrawOnChain")

	if len(datas) == 0 {
		log.Debug("Emtpy Withdraw data")
		return nil
	}

	// by chain withdraws map
	byChain := make(map[string][]withdrawOnChainData)

	for _, data := range datas {
		chain := data.Data.Chain
		if _, ok := byChain[chain]; !ok {
			byChain[chain] = make([]withdrawOnChainData, 0)
		}
		byChain[chain] = append(byChain[chain], data)
	}

	// process withdraw for same chain
	for chain, datas := range byChain {
		err := processWithdrawOnChainByNetwork(ctx, chain, datas)
		if err != nil {
			log.WithError(err).
				WithField("Chain", chain).
				Error("Failed to processWithdrawOnChainNetwork")
			continue
		}
	}

	return nil
}

func processWithdrawOnChainByNetwork(ctx context.Context, chain string, datas []withdrawOnChainData) error {
	log := logger.Logger(ctx).WithField("Method", "Accounting.processWithdrawOnChainByNetwork")
	db := appcontext.Database(ctx)

	if len(chain) == 0 {
		log.Error("Invalid chain")
		return ErrProcessingWithdraw
	}
	if len(datas) == 0 {
		log.Debug("Emtpy Withdraw data")
		return nil
	}

	// Acquire Lock
	lock, err := cache.LockBatchNetwork(ctx, chain)
	if err != nil {
		log.WithError(err).
			Error("Failed to lock batchNetwork")
		return ErrProcessingWithdraw
	}
	defer lock.Unlock()

	var canceled []model.WithdrawID

	// within a db transaction
	err = db.Transaction(func(db bank.Database) error {

		var IDs []model.WithdrawID
		withdrawPubkeyMap := make(map[model.WithdrawID]string)
		for _, data := range datas {
			// check if public key is valid
			if len(data.Data.PublicKey) == 0 {
				log.Error("Invalid Withdraw PublicKey")
				canceled = append(canceled, data.Withdraw.ID)
				continue
			}

			// store withdrawID publicKey
			withdrawPubkeyMap[data.Withdraw.ID] = data.Data.PublicKey

			// check if withdraw amount is valid
			if data.Withdraw.Amount == nil || *data.Withdraw.Amount <= 0.0 {
				log.Error("Invalid Withdraw Amount")
				canceled = append(canceled, data.Withdraw.ID)
				continue
			}

			// change to status processing
			_, err := database.AddWithdrawInfo(db, data.Withdraw.ID, model.WithdrawStatusProcessing, "{}")
			if err != nil {
				log.WithError(err).
					Error("Failed to AddWithdrawInfo")

				canceled = append(canceled, IDs...)
				continue
			}

			IDs = append(IDs, data.Withdraw.ID)
		}

		var batchOffset int
		for len(IDs) > 0 {
			// create new batch regarding batchOffset
			batchInfo, err := findOrCreateBatchInfo(db, chain, batchOffset)
			if err != nil {
				log.WithError(err).
					Error("Failed to findOrCreateBatchInfo")
				return ErrProcessingWithdraw
			}

			// get capacity of current batch
			count, capacity, withdrawIDs, err := batchWithdrawCount(db, batchInfo.BatchID)
			if err != nil {
				log.WithError(err).
					Error("Failed to batchWithdrawCount")
				return ErrProcessingWithdraw
			}

			if count == capacity {
				// seek to next batch
				batchOffset++
				continue
			}

			addressMap := make(map[string]model.WithdrawID)
			for _, withdrawID := range withdrawIDs {
				wt, err := database.GetWithdrawTargetByWithdrawID(db, withdrawID)
				if err != nil {
					log.WithError(err).
						Error("GetWithdrawTargetByWithdrawID Failed")
					return ErrProcessingWithdraw
				}
				data, err := wt.OnChainData()
				if err != nil {
					log.WithError(err).
						Error("WithdrawTarget OnChainData Failed")
					return ErrProcessingWithdraw
				}
				// mark address as used
				addressMap[data.PublicKey] = withdrawID
			}

			// get all batch IDs
			batchIDs := IDs[:]
			{
				remaining := capacity - count
				if len(IDs) <= remaining {
					// all remaining fits in current batch
					IDs = nil // stop loop
				} else {
					// truncate IDs with remaining batch capacity
					batchIDs, IDs = IDs[:remaining], IDs[remaining:] // update batchIDs & IDs
				}
			}

			// find & remove witdraw from batch with same PublicKey
			batchCopy := make([]model.WithdrawID, len(batchIDs))
			copy(batchCopy, batchIDs)
			for i, batchID := range batchCopy {
				pubKey := withdrawPubkeyMap[batchID]
				if _, exists := addressMap[pubKey]; exists {
					batchIDs = removeWithdraw(batchIDs, i)            // remove from current batch
					IDs = append([]model.WithdrawID{batchID}, IDs...) // prepend for next batch
					continue
				}
			}

			// Add witdraws to batch
			if len(batchIDs) > 0 {
				// append batchIds to current batch
				err = database.AddWithdrawToBatch(db, batchInfo.BatchID, batchIDs...)
				if err != nil {
					canceled = append(canceled, batchIDs...)
					log.WithError(err).
						Error("Failed to AddWithdrawToBatch")
					return ErrProcessingWithdraw
				}
			}

			if len(IDs) > 0 {
				batchOffset++ // increment to get new batch in next step
			}
		}

		return nil
	})

	// update all canceled withdraws
	for _, ID := range canceled {
		_, err := database.AddWithdrawInfo(db, ID, model.WithdrawStatusCanceled, "{}")
		if err != nil {
			log.WithError(err).Error("failed to cancelWithdraw")
			continue
		}
	}

	if err != nil {
		return ErrProcessingWithdraw
	}

	return nil
}

func batchWithdrawCount(db bank.Database, batchID model.BatchID) (int, int, []model.WithdrawID, error) {
	batch, err := database.GetBatch(db, batchID)
	if err != nil {
		return 0, 0, nil, err
	}
	withdraws, err := database.GetBatchWithdraws(db, batch.ID)
	if err != nil {
		return 0, 0, nil, err
	}

	return len(withdraws), int(batch.Capacity), withdraws, nil
}

func findOrCreateBatchInfo(db bank.Database, chain string, batchOffset int) (model.BatchInfo, error) {
	network := model.BatchNetwork(chain)
	batchCreated, err := database.GetLastBatchInfoByStatusAndNetwork(db, model.BatchStatusCreated, network)
	if err != nil {
		return model.BatchInfo{}, err
	}

	if len(batchCreated) > batchOffset {
		return batchCreated[batchOffset], nil
	}

	// create BatchInfo if not exists
	batch, err := database.AddBatch(db, network, model.BatchData(""))
	if err != nil {
		return model.BatchInfo{}, err
	}

	if err != nil {
		return model.BatchInfo{}, err
	}
	batchInfo, err := database.AddBatchInfo(db, batch.ID, model.BatchStatusCreated, model.BatchInfoCrypto, "{}")
	if err != nil {
		return model.BatchInfo{}, err
	}

	return batchInfo, nil
}

func removeWithdraw(s []model.WithdrawID, i int) []model.WithdrawID {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func OnCryptoValidateWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnCryptoWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CryptoValidateWithdraw
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			operation, err := CryptoValidateWithdraw(ctx, request.AuthInfo, request.ID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("CryptoWithdraw succeeded")

			// create & return response
			return &operation, nil
		})
}
