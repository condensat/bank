package services

import (
	"errors"
	"fmt"
	"net/http"

	"git.condensat.tech/bank/api/sessions"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	accounting "git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/wallet/client"

	"github.com/sirupsen/logrus"
)

var (
	ErrWalletChainNotFoundError = errors.New("Chain Not Found")
)

type WalletService int

// WalletNextDepositRequest holds args for accounting requests
type WalletNextDepositRequest struct {
	SessionArgs
	AccountID string `json:"accountId"`
}

// WalletNextDepositResponse holds args for accounting requests
type WalletNextDepositResponse struct {
	Currency        string `json:"currency"`
	DisplayCurrency string `json:"displayCurrency"`
	PublicAddress   string `json:"publicAddress"`
	URL             string `json:"url"`
}

// WalletService operation return deposit address for account
func (p *WalletService) NextDeposit(r *http.Request, request *WalletNextDepositRequest, reply *WalletNextDepositResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "WalletService.NextDeposit")
	log = GetServiceRequestLog(log, r, "Wallet", "NextDeposit")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = getSessionCookie(r)
	sessionID := sessions.SessionID(request.SessionID)
	userID := session.UserSession(ctx, sessionID)
	if !sessions.IsUserValid(userID) {
		log.Error("Invalid userSession")
		return sessions.ErrInvalidSessionID
	}
	log = log.WithFields(logrus.Fields{
		"SessionID": sessionID,
		"UserID":    userID,
	})

	sID := appcontext.SecureID(ctx)
	accountID, err := sID.FromSecureID("account", sID.Parse(request.AccountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("Wrong AccountID")
		return sessions.ErrInternalError
	}

	account, err := accounting.AccountInfo(ctx, uint64(accountID))
	if err != nil {
		log.WithError(err).Error("AccountInfo failed")
		return err
	}
	if !account.Currency.Crypto {
		log.WithField("AccountID", request.AccountID).
			Error("Non Crypto Account")
		return sessions.ErrInternalError
	}
	chain, err := getChainFromCurrencyName(account.Currency.Crypto, account.Currency.Name)
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("getChainFromCurrencyName failed")
		return sessions.ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"Chain":     chain,
		"AccountID": accountID,
	})

	addr, err := client.CryptoAddressNextDeposit(ctx, chain, uint64(accountID))
	if err != nil {
		log.WithError(err).
			Error("CryptoAddressNextDeposit Failed")
		return ErrServiceInternalError
	}

	// Reply
	protocol, err := getProtocolFromCurrencyName(account.Currency.Crypto, account.Currency.Name)
	*reply = WalletNextDepositResponse{
		Currency:        account.Currency.Name,
		DisplayCurrency: account.Currency.DisplayName,
		PublicAddress:   addr.PublicAddress,
		URL:             fmt.Sprintf("%s:%s", protocol, addr.PublicAddress),
	}
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			WithField("CurrencyName", account.Currency.Name).
			Error("getProtocolFromCurrencyName failed")
		return sessions.ErrInternalError
	}

	log.WithFields(logrus.Fields{
		"CurrencyName":  reply.Currency,
		"PublicAddress": reply.PublicAddress,
		"Url":           reply.URL,
	}).Info("CryptoAddressNextDeposit")

	return nil
}

func getChainFromCurrencyName(isCrypto bool, currencyName string) (string, error) {
	switch currencyName {
	case "BTC":
		return "bitcoin-mainnet", nil
	case "TBTC":
		return "bitcoin-testnet", nil
	case "LBTC":
		return "liquid-mainnet", nil

	default:
		if isCrypto {
			return "liquid-mainnet", nil
		}
		return "", ErrWalletChainNotFoundError
	}
}

func getProtocolFromCurrencyName(isCrypto bool, currencyName string) (string, error) {
	switch currencyName {
	case "BTC":
		return "bitcoin", nil
	case "TBTC":
		return "bitcoin", nil
	case "LBTC":
		return "liquid", nil

	default:
		if isCrypto {
			return "liquid", nil
		}
		return "", ErrWalletChainNotFoundError
	}
}