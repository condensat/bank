package handlers

import (
	"context"
	"errors"
	"math"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"git.condensat.tech/bank/accounting/common"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"github.com/sirupsen/logrus"
)

const (
	BankWitdrawAccountName = model.AccountName("withdraw")
)

func AccountTransferWithdrawCrypto(ctx context.Context, withdraw common.AccountTransferWithdrawCrypto) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountTransferWithdrawCrypto")
	db := appcontext.Database(ctx)

	bankAccountID, err := getBankWithdrawAccount(ctx, withdraw.Source.Currency)
	if err != nil {
		log.WithError(err).
			Error("Invalid BankAccount")
		return common.AccountTransfer{}, database.ErrInvalidAccountID
	}

	log = log.WithFields(logrus.Fields{
		"BankAccountId": bankAccountID,
		"Currency":      withdraw.Source.Currency,
	})

	// get ticker precision to convert back in BTC precision (for RPC)
	tickerPrecision := -1 // no ticker precison if not crypto
	currency, err := database.GetCurrencyByName(db, model.CurrencyName(withdraw.Source.Currency))
	if err != nil {
		return common.AccountTransfer{}, err
	}
	asset, _ := database.GetAssetByCurrencyName(db, currency.Name)

	isAsset := currency.IsCrypto() && currency.GetType() == 2 && asset.ID > 0
	if currency.IsCrypto() {
		tickerPrecision = 8 // BTC precision
	}
	if isAsset {
		tickerPrecision = 0
		if assetInfo, err := database.GetAssetInfo(db, asset.ID); err == nil {
			tickerPrecision = int(assetInfo.Precision)
		}

		if currency.Name == "LBTC" {
			tickerPrecision = 8 // BTC precision
		}
	}

	feeCurrencyName := getFeeCurrency(string(currency.Name), isAsset)

	feeBankAccountID, err := getBankWithdrawAccount(ctx, feeCurrencyName)
	if err != nil {
		log.WithError(err).
			Error("Invalid Fee BankAccount")
		return common.AccountTransfer{}, err
	}

	// convert amount in BTC precision
	amount := convertAssetAmountToBitcoin(withdraw.Source.Amount, tickerPrecision)
	if amount <= 0.0 {
		return common.AccountTransfer{}, database.ErrInvalidWithdrawAmount
	}

	log.WithFields(logrus.Fields{
		"IsAsset":         isAsset,
		"Asset":           asset,
		"Currency":        withdraw.Source.Currency,
		"CurrencyInfo":    currency,
		"BitcoinAmount":   amount,
		"TickerPrecision": tickerPrecision,
		"AssetAmount":     withdraw.Source.Amount,
	}).Debug("Asset to Bitcoin precision")

	batchMode := model.BatchModeNormal
	if len(withdraw.BatchMode) > 0 {
		batchMode = model.BatchMode(withdraw.BatchMode)
	}

	var result common.AccountTransfer
	// Database Query
	err = db.Transaction(func(db bank.Database) error {

		// Create Witdraw for batch
		w, err := database.AddWithdraw(db,
			model.AccountID(withdraw.Source.AccountID),
			model.AccountID(bankAccountID),
			model.Float(amount), batchMode,
			"{}",
		)
		if err != nil {
			log.WithError(err).
				Error("AddWithdraw failed")
			return err
		}
		_, err = database.AddWithdrawInfo(db, w.ID, model.WithdrawStatusCreated, "{}")
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawInfo failed")
			return err
		}

		wt := model.FromOnChainData(w.ID, withdraw.Crypto.Chain, model.WithdrawTargetOnChainData{
			WithdrawTargetCryptoData: model.WithdrawTargetCryptoData{
				PublicKey: withdraw.Crypto.PublicKey,
			},
		})

		_, err = database.AddWithdrawTarget(db, w.ID, wt.Type, wt.Data)
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawTarget failed")
			return err
		}

		referenceID := uint64(w.ID)

		// get fee informations
		feeInfo, err := database.GetFeeInfo(db, model.CurrencyName(feeCurrencyName))
		if err != nil {
			log.WithError(err).
				Error("GetFeeInfo failed")
			return err
		}
		if !feeInfo.IsValid() {
			log.Error("Invalid FeeInfo")
			return errors.New("Invalid FeeInfo")
		}

		feeAmount := feeInfo.Compute(model.Float(amount))
		feeUserAccount := withdraw.Source.AccountID
		if feeCurrencyName != withdraw.Source.Currency {
			// if fee is not in the same currency (ie asset without quote)
			// take the minimum fee of the currency fee
			feeAmount = feeInfo.Minimum

			// get feeUserAccoiunt from user
			userAccount, err := database.GetAccountByID(db, model.AccountID(withdraw.Source.AccountID))
			if err != nil {
				log.WithError(err).
					Error("GetAccountByID failed")
				return err
			}
			// get user account for currency fee
			accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, userAccount.UserID, model.CurrencyName(feeCurrencyName), database.AccountNameDefault)
			if err != nil {
				return errors.New("GetAccountsByUserAndCurrencyAndName failed")
			}
			if len(accounts) == 0 {
				return database.ErrAccountNotFound
			}
			// use first default account
			account := accounts[0]
			feeUserAccount = uint64(account.ID)
		}

		// Transfert fees from account to bankAccount
		timestamp := common.Timestamp()
		result, err = AccountTransferWithDatabase(ctx, db, common.AccountTransfer{
			Source: common.AccountEntry{
				AccountID: feeUserAccount,

				OperationType:    string(model.OperationTypeTransferFee),
				SynchroneousType: "sync",
				ReferenceID:      referenceID,

				Timestamp: timestamp,
				Amount:    float64(-feeAmount),

				Currency: feeCurrencyName,
			},
			Destination: common.AccountEntry{
				AccountID: uint64(feeBankAccountID),

				OperationType:    string(model.OperationTypeTransferFee),
				SynchroneousType: "sync",
				ReferenceID:      referenceID,

				Timestamp: timestamp,
				Amount:    float64(feeAmount),

				Currency: feeCurrencyName,
			},
		})
		if err != nil {
			log.WithError(err).
				Error("AccountTransfer fee failed")
			return err
		}

		// Transfert amount from account to bank account
		result, err = AccountTransferWithDatabase(ctx, db, common.AccountTransfer{
			Source: withdraw.Source,
			Destination: common.AccountEntry{
				AccountID: uint64(bankAccountID),

				OperationType:    withdraw.Source.OperationType,
				SynchroneousType: "async-start",
				ReferenceID:      referenceID,

				Timestamp: common.Timestamp(),
				Amount:    amount,

				Label: withdraw.Source.Label,

				LockAmount: amount,
				Currency:   withdraw.Source.Currency,
			},
		})
		if err != nil {
			log.WithError(err).
				Error("AccountTransfer failed")
			return err
		}

		log.Debug("AccountWithdraw created")

		return nil
	})
	if err != nil {
		return common.AccountTransfer{}, err
	}

	return result, err
}

func AccountTransferWithdrawFiat(ctx context.Context, withdraw common.AccountTransferWithdrawFiat) (common.AccountTransfer, error) {
	log := logger.Logger(ctx).WithField("Method", "accounting.AccountTransferWithdrawFiat")
	db := appcontext.Database(ctx)

	var result common.AccountTransfer

	// Sanity checks
	if withdraw.UserID == 0 {
		return result, errors.New("Invalid UserID")
	}

	log.WithField("userID", withdraw.UserID)

	if withdraw.Source.Amount <= 0.0 {
		return result, errors.New("Amount can't be nul or negative")
	}

	if withdraw.Source.Amount < minAmountFiatWithdraw {
		return result, errors.New("Amount is below the minimum required")
	}

	if withdraw.Source.LockAmount != 0.0 {
		return result, errors.New("LockAmount must be 0")
	}

	// check that IBAN is in the correct format
	validIban, err := withdraw.Sepa.IBAN.Valid()
	if err != nil {
		return result, err
	}
	if !validIban {
		return result, errors.New("Provided iban invalid format")
	}

	// check that Bic is correct format
	validBic, err := withdraw.Sepa.BIC.Valid()
	if err != nil {
		return result, err
	}
	if !validBic {
		return result, errors.New("Provided bic invalid format")
	}

	// check operation type is fiat_withdraw
	if withdraw.Source.OperationType != string(model.OperationTypeTransfer) {
		return result, errors.New("Invalid Operation type")
	}

	// check sync is sync
	if withdraw.Source.SynchroneousType != string(model.SynchroneousTypeSync) {
		return result, errors.New("Invalid Sync type")
	}

	bankAccountID, err := getBankWithdrawAccount(ctx, withdraw.Source.Currency)
	if err != nil {
		log.WithError(err).
			Error("Invalid BankAccount")
		return result, database.ErrInvalidAccountID
	}

	log = log.WithFields(logrus.Fields{
		"BankAccountId": bankAccountID,
		"Currency":      withdraw.Source.Currency,
		"Amount":        withdraw.Source.Amount,
	})

	currency, err := database.GetCurrencyByName(db, model.CurrencyName(withdraw.Source.Currency))
	if err != nil {
		return result, err
	}

	// Is the currency fiat?
	if currency.IsCrypto() {
		return result, errors.New("Currency is not fiat")
	}

	// Round up to currency.DisplayPrecision
	rounding := math.Pow10(int(*currency.Precision))
	withdraw.Source.Amount = math.Floor(withdraw.Source.Amount*rounding) / rounding

	// Compute fees amount and find bank account to pay to
	feeInfo, err := database.GetFeeInfo(db, currency.Name)
	if err != nil {
		log.WithError(err).
			Error("GetFeeInfo failed")
		return result, err
	}
	if !feeInfo.IsValid() {
		log.Error("Invalid FeeInfo")
		return result, err
	}

	feeAmount := feeInfo.Compute(model.Float(withdraw.Source.Amount))

	if feeAmount < 0 {
		return result, errors.New("Negative fee amount are not allowed")
	}

	if feeAmount < feeInfo.Minimum {
		feeAmount = feeInfo.Minimum
	}

	log = log.WithField("feeAmount", feeAmount)

	feeBankAccountID, err := getBankWithdrawAccount(ctx, withdraw.Source.Currency)
	if err != nil {
		return result, errors.New("Can't get bank account id")
	}

	// Get AccountID with UserID only if no accountId provided
	if withdraw.Source.AccountID == 0 {
		accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, model.UserID(withdraw.UserID), model.CurrencyName(currency.Name), model.AccountName("*"))
		if err != nil || len(accounts) == 0 {
			return result, errors.New("Accounts not found")
		}

		for _, account := range accounts {
			// get account info
			accountInfo, err := AccountInfo(ctx, uint64(account.ID))
			if err != nil {
				return result, err
			}
			if accountInfo.Status != "normal" {
				continue
			}

			// Check available balance too
			if (withdraw.Source.Amount + float64(feeAmount)) > accountInfo.Balance {
				continue
			}

			// We found a suitable account
			withdraw.Source.AccountID = uint64(account.ID)
			break
		}
	}

	if withdraw.Source.AccountID == 0 {
		return result, errors.New("Can't find an account that allows withdraw for this user and currency")
	}

	// Look for the sepa with userID and IBAN
	sepaUser, err := database.GetSepaByUserAndIban(db, model.UserID(withdraw.UserID), model.Iban(withdraw.Sepa.IBAN))
	if err != nil && err != database.ErrSepaNotFound {
		return result, err
	}

	if sepaUser.ID == 0 {

		// if sepa is not registered, add it to db
		sepaUser, err = database.CreateSepa(db, model.FiatSepaInfo{
			UserID: model.UserID(withdraw.UserID),
			IBAN:   model.Iban(withdraw.Sepa.IBAN),
			BIC:    model.Bic(withdraw.Sepa.BIC),
			Label:  model.String(withdraw.Sepa.Label),
		})
		if err != nil {
			return result, err
		}

	} else {

		// Is there a fiatoperation for this sepa AND this user?
		fiatOperation, err := database.FindFiatWithdrawalPendingForUserAndSepa(db, model.UserID(withdraw.UserID), sepaUser.ID)
		if err != nil {
			return result, err
		}

		// stop if there's already 1 or more pending withdrawal
		switch len := len(fiatOperation); len {
		case 0:
			break
		case 1:
			return result, errors.New("Already a pending withdrawal for this user and sepa")
		default:
			return result, errors.New("Multiple pending withdrawals for this user and sepa")
		}
	}

	batchMode := model.BatchModeNormal
	if len(withdraw.BatchMode) > 0 {
		batchMode = model.BatchMode(withdraw.BatchMode)
	}

	// Database Query
	err = db.Transaction(func(db bank.Database) error {

		// Create Witdraw for batch
		w, err := database.AddWithdraw(db,
			model.AccountID(withdraw.Source.AccountID),
			model.AccountID(bankAccountID),
			model.Float(withdraw.Source.Amount), batchMode,
			"{}",
		)
		if err != nil {
			log.WithError(err).
				Error("AddWithdraw failed")
			return err
		}
		_, err = database.AddWithdrawInfo(db, w.ID, model.WithdrawStatusCreated, "{}")
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawInfo failed")
			return err
		}

		wt := model.FromSepaData(w.ID, model.WithdrawTargetSepaData{
			BIC:  string(withdraw.Sepa.BIC),
			IBAN: string(withdraw.Sepa.IBAN),
		},
		)

		_, err = database.AddWithdrawTarget(db, w.ID, wt.Type, wt.Data)
		if err != nil {
			log.WithError(err).
				Error("AddWithdrawTarget failed")
			return err
		}

		referenceID := uint64(w.ID)

		feeSource := common.AccountEntry{
			AccountID: withdraw.Source.AccountID,

			OperationType:    string(model.OperationTypeTransferFee),
			SynchroneousType: "sync",

			Currency: withdraw.Source.Currency,
		}
		feeDestination := common.AccountEntry{
			AccountID: uint64(feeBankAccountID),

			OperationType:    string(model.OperationTypeTransferFee),
			SynchroneousType: "sync",
			ReferenceID:      withdraw.Source.ReferenceID,

			Timestamp: common.Timestamp(),

			Amount: float64(feeAmount),

			Currency: withdraw.Source.Currency,
		}
		_, err = AccountTransferWithDatabase(ctx, db, common.AccountTransfer{
			Source:      feeSource,
			Destination: feeDestination,
		})
		if err != nil {
			log.WithError(err).Error("AccountTransferWithDatabase failed")
			return errors.New("transfer fee operation failed")
		}

		// Transfert amount from account to bank account
		result, err = AccountTransferWithDatabase(ctx, db, common.AccountTransfer{
			Source: withdraw.Source,
			Destination: common.AccountEntry{
				AccountID: uint64(bankAccountID),

				OperationType:    withdraw.Source.OperationType,
				SynchroneousType: "async-start",
				ReferenceID:      referenceID,

				Timestamp: common.Timestamp(),
				Amount:    withdraw.Source.Amount,

				Label: withdraw.Source.Label,

				LockAmount: withdraw.Source.Amount,
				Currency:   withdraw.Source.Currency,
			},
		})
		if err != nil {
			log.WithError(err).
				Error("AccountTransfer failed")
			return err
		}

		log.Debug("AccountWithdraw created")

		return nil
	})
	if err != nil {
		return common.AccountTransfer{}, err
	}

	return result, err
}

func OnAccountTransferWithdrawCrypto(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountTransferWithdrawCrypto")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountTransferWithdrawCrypto
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := AccountTransferWithdrawCrypto(ctx, request)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"AccountID": request.Source.AccountID,
					}).Errorf("Failed to AccountTransferWithdrawCrypto")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}

func OnAccountTransferWithdrawFiat(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnAccountTransferWithdrawFiat")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AccountTransferWithdrawFiat
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			response, err := AccountTransferWithdrawFiat(ctx, request)
			if err != nil {
				log.WithError(err).
					WithFields(logrus.Fields{
						"AccountID": request.Source.AccountID,
					}).Errorf("Failed to AccountTransferWithdraw")
				return nil, cache.ErrInternalError
			}

			// return response
			return &response, nil
		})
}

func getBankWithdrawAccount(ctx context.Context, currency string) (model.AccountID, error) {
	bankUser := common.BankUserFromContext(ctx)
	if bankUser.ID == 0 {
		return 0, database.ErrInvalidUserID
	}

	db := appcontext.Database(ctx)
	currencyName := model.CurrencyName(currency)
	if !database.AccountsExists(db, bankUser.ID, currencyName, BankWitdrawAccountName) {
		result, err := AccountCreate(ctx, uint64(bankUser.ID), common.AccountInfo{
			UserID: uint64(bankUser.ID),
			Name:   string(BankWitdrawAccountName),
			Currency: common.CurrencyInfo{
				Name: currency,
			},
		})
		if err != nil {
			return 0, err
		}

		_, err = AccountSetStatus(ctx, result.AccountID, model.AccountStatusNormal.String())
		if err != nil {
			return 0, err
		}
		return model.AccountID(result.AccountID), err
	}

	accounts, err := database.GetAccountsByUserAndCurrencyAndName(db, bankUser.ID, model.CurrencyName(currencyName), BankWitdrawAccountName)
	if err != nil {
		return 0, err
	}

	if len(accounts) == 0 {
		return 0, database.ErrAccountNotFound
	}
	account := accounts[0]
	if account.ID == 0 {
		return 0, database.ErrInvalidAccountID
	}

	return account.ID, nil
}

func getFeeCurrency(currency string, isAsset bool) string {
	if !isAsset {
		return currency
	}

	switch currency {
	case "USDt":
		fallthrough
	case "LCAD":
		return currency

	default:
		return "LBTC"
	}
}
