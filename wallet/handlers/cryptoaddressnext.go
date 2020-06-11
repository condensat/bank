package handlers

import (
	"context"
	"errors"
	"fmt"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/messaging"

	"github.com/shengdoushi/base58"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidChain     = errors.New("Invalid Chain")
	ErrInvalidAccountID = errors.New("Invalid AccountID")
	ErrGenAddress       = errors.New("Gen Address Error")
)

func CryptoAddressNextDeposit(ctx context.Context, address common.CryptoAddress) (common.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.CryptoAddressNextDeposit")
	var result common.CryptoAddress

	chainHandler := ChainHandlerFromContext(ctx)
	if chainHandler == nil {
		log.Error("Failed to ChainHandlerFromContext")
		return result, ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"Chain":     address.Chain,
		"AccountID": address.AccountID,
	})

	if len(address.Chain) == 0 {
		log.WithError(ErrInvalidChain).
			Debug("AddressNext Failed")
		return result, ErrInvalidChain
	}
	if address.AccountID == 0 {
		log.WithError(ErrInvalidAccountID).
			Debug("AddressNext Failed")
		return result, ErrInvalidAccountID
	}

	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {

		chain := model.String(address.Chain)
		accountID := model.AccountID(address.AccountID)

		addresses, err := database.AllUnusedAccountCryptoAddresses(db, accountID)
		if err != nil {
			log.WithError(err).
				Error("Failed to AllUnusedAccountCryptoAddresses")
			return err
		}

		// return last unised address
		if len(addresses) > 0 {
			addr := addresses[len(addresses)-1]

			log.Debug("Found unused deposit address")

			result = common.CryptoAddress{
				CryptoAddressID: uint64(addr.ID),
				Chain:           string(addr.Chain),
				AccountID:       uint64(addr.AccountID),
				PublicAddress:   string(addr.PublicAddress),
				Unconfidential:  string(addr.Unconfidential),
			}
			return nil
		}

		addr, err := txNewCryptoAddress(ctx, db, chainHandler, chain, accountID)
		if err != nil {
			log.WithError(err).
				Error("Failed to txNewCryptoAddress")
			return err
		}

		result = common.CryptoAddress{
			CryptoAddressID: uint64(addr.ID),
			Chain:           string(addr.Chain),
			AccountID:       uint64(addr.AccountID),
			PublicAddress:   string(addr.PublicAddress),
			Unconfidential:  string(addr.Unconfidential),
		}

		return nil
	})

	if err == nil {
		log.WithField("PublicAddress", result.PublicAddress).
			Debug("Next deposit publicAddress")
	}

	return result, err
}

func OnCryptoAddressNextDeposit(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnCryptoAddressNextDeposit")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.CryptoAddress
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"Chain":     request.Chain,
				"AccountID": request.AccountID,
			})

			nextDeposit, err := CryptoAddressNextDeposit(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoAddressNextDeposit")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"PublicAddress": nextDeposit.PublicAddress,
			})

			log.Info("Next Deposit Address")

			// create & return response
			return &common.CryptoAddress{
				CryptoAddressID: nextDeposit.CryptoAddressID,
				Chain:           request.Chain,
				AccountID:       request.AccountID,
				PublicAddress:   nextDeposit.PublicAddress,
				Unconfidential:  nextDeposit.Unconfidential,
			}, nil
		})
}

func genAccountLabelFromAccountID(accountID model.AccountID) string {
	// create account label from accountID
	accountHash := fmt.Sprintf("bank.account:%d", accountID)
	return base58.Encode([]byte(accountHash), base58.BitcoinAlphabet)
}