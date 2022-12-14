package handlers

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func CryptoAddressNewDeposit(ctx context.Context, address common.CryptoAddress) (common.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.CryptoAddressNewDeposit")
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

		addr, err := txNewCryptoAddress(ctx, db, chainHandler, chain, accountID, address.IgnoreAccounting)
		if err != nil {
			log.WithError(err).
				Error("Failed to txNewCryptoAddress")
			return err
		}

		result = convertCryptoAddress(addr)

		return nil
	})

	if err == nil {
		log.WithField("PublicAddress", result.PublicAddress).
			Debug("New deposit publicAddress")
	}

	return result, err
}

func OnCryptoAddressNewDeposit(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.OnCryptoAddressNewDeposit")
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

			newDeposit, err := CryptoAddressNewDeposit(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoAddressNewsDeposit")
				return nil, cache.ErrInternalError
			}

			log = log.WithFields(logrus.Fields{
				"PublicAddress": newDeposit.PublicAddress,
			})

			log.Info("New Deposit Address")

			// create & return response
			return &common.CryptoAddress{
				CryptoAddressID: newDeposit.CryptoAddressID,
				Chain:           request.Chain,
				AccountID:       request.AccountID,
				PublicAddress:   newDeposit.PublicAddress,
				Unconfidential:  newDeposit.Unconfidential,
			}, nil
		})
}

func txNewCryptoAddress(ctx context.Context, db bank.Database, chainHandler ChainHandler, chain model.String, accountID model.AccountID, ignoreAccounting bool) (model.CryptoAddress, error) {
	var result model.CryptoAddress
	var err error
	errCall := cache.ExecuteSingleCall(ctx, "txNewCryptoAddress", func(ctx context.Context) error {

		switch common.CryptoModeFromContext(ctx) {
		case common.CryptoModeCryptoSsm:
			result, err = txNewCryptoAddressSsm(ctx, db, chainHandler, chain, accountID, ignoreAccounting)

		default:
			result, err = txNewCryptoAddressFullNode(ctx, db, chainHandler, chain, accountID, ignoreAccounting)
		}

		return err
	})

	if errCall != nil {
		return model.CryptoAddress{}, err
	}

	return result, err
}

func txNewCryptoAddressFullNode(ctx context.Context, db bank.Database, chainHandler ChainHandler, chain model.String, accountID model.AccountID, ignoreAccounting bool) (model.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.txNewCryptoAddressFullNode")
	account := genAccountLabelFromAccountID(accountID)
	publicAddress, err := chainHandler.GetNewAddress(ctx, string(chain), account)
	if err != nil {
		log.WithError(err).
			WithField("Chain", chain).
			Error("Failed to GetNewAddress")
		return model.CryptoAddress{}, ErrGenAddress
	}

	info, err := chainHandler.GetAddressInfo(ctx, string(chain), publicAddress)
	if err != nil {
		log.WithError(err).
			Error("Failed to GetAddressInfo")
		return model.CryptoAddress{}, ErrGenAddress
	}

	addr, err := database.AddOrUpdateCryptoAddress(db, model.CryptoAddress{
		Chain:            chain,
		AccountID:        accountID,
		PublicAddress:    model.String(publicAddress),
		Unconfidential:   model.String(info.Unconfidential),
		IgnoreAccounting: ignoreAccounting,
	})
	if err != nil {
		log.WithError(err).
			Error("Failed to AddOrUpdateCryptoAddress")
		return model.CryptoAddress{}, err
	}

	return addr, nil
}

func txNewCryptoAddressSsm(ctx context.Context, db bank.Database, chainHandler ChainHandler, chain model.String, accountID model.AccountID, ignoreAccounting bool) (model.CryptoAddress, error) {
	log := logger.Logger(ctx).WithField("Method", "wallet.txNewCryptoAddressSsm")
	account := genAccountLabelFromAccountID(accountID)

	ssmChain := convertToSsmChain(chain)
	if len(ssmChain) == 0 {
		log.WithField("Chain", chain).
			Error("Invalid ssm chain")
		return model.CryptoAddress{}, ErrGenAddress
	}

	fingerprint := getSsmFingerprintFromChain(ctx, ssmChain)
	if len(fingerprint) == 0 {
		log.
			Error("Invalid fingerprint")
		return model.CryptoAddress{}, ErrGenAddress
	}

	var result model.CryptoAddress
	// already within a db transaction
	err := func(db bank.Database) error {

		ssmAddressID, err := database.NextSsmAddressID(db, ssmChain, fingerprint)
		if err != nil {
			log.WithError(err).
				Error("Failed to NextSsmAddressID")
			return ErrGenAddress
		}

		ssmAddress, err := database.GetSsmAddress(db, ssmAddressID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetSsmAddress")
			return ErrGenAddress
		}

		_, err = database.UpdateSsmAddressState(db, ssmAddressID, model.SsmAddressStatusUsed)
		if err != nil {
			if err != nil {
				log.WithError(err).
					Error("Failed to GetSsmAddress")
				return ErrGenAddress
			}
		}

		publicAddress := string(ssmAddress.PublicAddress)
		pubkey := string(ssmAddress.ScriptPubkey)
		blindingKey := string(ssmAddress.BlindingKey)
		err = chainHandler.ImportAddress(ctx, string(chain), account, publicAddress, pubkey, blindingKey)
		if err != nil {
			log.WithError(err).
				Error("Failed to ImportAddress")
			return ErrGenAddress
		}

		info, err := chainHandler.GetAddressInfo(ctx, string(chain), publicAddress)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetAddressInfo")
			return ErrGenAddress
		}

		result, err = database.AddOrUpdateCryptoAddress(db, model.CryptoAddress{
			Chain:            chain,
			AccountID:        accountID,
			PublicAddress:    model.String(publicAddress),
			Unconfidential:   model.String(info.Unconfidential),
			IgnoreAccounting: ignoreAccounting,
		})
		if err != nil {
			log.WithError(err).
				Error("Failed to AddOrUpdateCryptoAddress")
			return err
		}

		return nil
	}(db)
	if err != nil {
		log.WithError(err).
			Error("Failed to generate new CryptoAddress with Ssm")
		return model.CryptoAddress{}, ErrGenAddress
	}

	return result, nil
}

func convertToSsmChain(chain model.String) model.SsmChain {
	switch chain {

	case "bitcoin-mainnet":
		return "bitcoin-main"

	case "bitcoin-testnet":
		return "bitcoin-test"

	case "liquid-mainnet":
		return "liquidv1"

	default:
		return ""
	}
}

func getSsmFingerprintFromChain(ctx context.Context, chain model.SsmChain) model.SsmFingerprint {
	ssmInfo := common.SsmDeviceInfoFromContext(ctx)
	if ssmInfo == nil {
		return ""
	}

	result, err := ssmInfo.Fingerprint(ctx, common.SsmChain(chain))
	if err != nil {
		return ""
	}
	return model.SsmFingerprint(result)
}
