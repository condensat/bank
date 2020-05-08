package handlers

import (
	"context"
	"fmt"
	"math"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/utils"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"git.condensat.tech/bank/accounting/common"

	"github.com/sirupsen/logrus"
)

func AccountInfo(ctx context.Context, accountID uint64) (common.AccountInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountInfo")

	log = log.WithFields(logrus.Fields{
		"AccountID": accountID,
	})

	var result common.AccountInfo
	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		account, err := database.GetAccountByID(db, model.AccountID(accountID))
		if err != nil {
			return err
		}

		result, err = txGetAccountInfo(db, account)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.WithError(err).
			Error("Failed to get AccountInfo")
		return common.AccountInfo{}, err
	}

	return result, nil
}

func txGetAccountInfo(db bank.Database, account model.Account) (common.AccountInfo, error) {
	currency, err := database.GetCurrencyByName(db, account.CurrencyName)
	if err != nil {
		return common.AccountInfo{}, err
	}
	accountState, err := database.GetAccountStatusByAccountID(db, account.ID)
	if err != nil {
		return common.AccountInfo{}, err
	}

	last, err := database.GetLastAccountOperation(db, account.ID)
	if err != nil {
		return common.AccountInfo{}, err
	}

	var balance float64
	var totalLocked float64
	if last.IsValid() {
		balance = float64(*last.Balance)
		totalLocked = float64(*last.TotalLocked)
	}

	asset, _ := database.GetAssetByCurrencyName(db, currency.Name)

	isAsset := currency.IsCrypto() && asset.ID > 0

	currencyName := string(currency.Name)
	displayPrecision := currency.DisplayPrecision()
	tickerPrecision := -1 // no ticker precison if not crypto
	if currency.IsCrypto() {
		tickerPrecision = 8 // BTC precision
	}
	if isAsset {
		currencyName = shortAssetHash(string(asset.Hash))
		displayPrecision = 0
		tickerPrecision = 0
		if assetInfo, err := database.GetAssetInfo(db, asset.ID); err == nil {
			tickerPrecision = int(assetInfo.Precision)
			currencyName = assetInfo.Name
		}
	}

	return common.AccountInfo{
		Timestamp: last.Timestamp,
		AccountID: uint64(account.ID),
		Currency: common.CurrencyInfo{
			Name:             currencyName,
			Crypto:           currency.IsCrypto(),
			Asset:            isAsset,
			DisplayPrecision: uint(displayPrecision),
		},
		Name:        string(account.Name),
		Status:      string(accountState.State),
		Balance:     convertAssetAmount(float64(balance), tickerPrecision),
		TotalLocked: convertAssetAmount(float64(totalLocked), tickerPrecision),
	}, nil
}

func OnAccountInfo(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountInfo")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
			})

			info, err := AccountInfo(ctx, request.AccountID)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to get AccountInfo")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &info, nil
		})
}

func shortAssetHash(hash string) string {
	const limit = 5
	if len(hash) <= 2*limit {
		return hash
	}
	return fmt.Sprintf("%s...%s", hash[:limit], hash[len(hash)-limit:])
}

func convertAssetAmount(amount float64, tickerPrecision int) float64 {
	if tickerPrecision < 0 {
		return amount
	}
	const btcPrecision = 8
	if tickerPrecision > btcPrecision {
		tickerPrecision = btcPrecision
	}
	amount *= math.Pow(10.0, float64(btcPrecision-tickerPrecision))

	return utils.ToFixed(amount, tickerPrecision)
}
