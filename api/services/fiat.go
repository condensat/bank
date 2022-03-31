package services

import (
	"net/http"

	"code.condensat.tech/bank/secureid"
	"git.condensat.tech/bank/accounting/client"
	accounting "git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/api/sessions"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"github.com/sirupsen/logrus"
)

type FiatService int

type FiatWithdrawRequest struct {
	SessionArgs
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Iban      string  `json:"iban"`
	Bic       string  `json:"bic"`
	SepaLabel string  `json:"sepaLabel,omitempty"`
	AccountID string  `json:"accountId,omitempty"`
}

type FiatWithdrawResponse struct {
	FiatWithdrawID string `json:"fiatWithdrawId"`
}

func (p *FiatService) Withdraw(r *http.Request, request *FiatWithdrawRequest, reply *FiatWithdrawResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "FiatService.Withdraw")
	log = GetServiceRequestLog(log, r, "Fiat", "Withdraw")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = GetSessionCookie(r)
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

	var accountId uint64
	if len(request.AccountID) != 0 {
		sIdValue, err := sID.FromSecureID("account", sID.Parse(request.AccountID))
		if err != nil {
			log.WithError(err).
				WithField("AccountID", request.AccountID).
				Error("Wrong AccountID")
			return sessions.ErrInternalError
		}

		accountId = uint64(sIdValue)

		account, err := accounting.AccountInfo(ctx, accountId)
		if err != nil {
			log.WithError(err).Error("AccountInfo failed")
			return err
		}
		if account.Status != "normal" {
			log.WithFields(logrus.Fields{
				"AccountID": request.AccountID,
				"Status":    account.Status,
			}).Error("Account status does not allow deposit")
			return ErrInvalidAccountID
		}
		if account.Currency.Crypto {
			log.WithField("AccountID", request.AccountID).
				Error("Crypto Account")
			return sessions.ErrInternalError
		}
		if account.Currency.Name != request.Currency {
			log.WithFields(logrus.Fields{
				"Withdraw currency": request.Currency,
				"Account currency":  account.Currency.Name,
			}).Error("Currency don't match with provided account to withdraw")
		}

	} else {
		accountId = uint64(0)
		log.Info("No accountId provided, inferring it later from userId and currency")
	}

	// Call internal API
	withdraw, err := client.AccountTransferWithdrawFiat(ctx, userID, accountId, request.Currency, request.Amount, "normal", request.Iban, request.Bic, request.SepaLabel)
	if err != nil {
		log.WithError(err).
			Error("FiatWithdraw failed")
		return ErrServiceInternalError
	}

	secureID, err := sID.ToSecureID("withdraw", secureid.Value(withdraw))
	if err != nil {
		log.WithError(err).
			Error("ToSecureID Failed")
		return ErrServiceInternalError
	}

	*reply = FiatWithdrawResponse{
		FiatWithdrawID: sID.ToString(secureID),
	}

	return nil
}

// WalletCancelWithdrawRequest holds args for wallet requests
type FiatCancelWithdrawRequest struct {
	SessionArgs
	WithdrawID string `json:"withdrawId"`
}

// WalletCancelWithdrawResponse holds args for wallet requests
type FiatCancelWithdrawResponse struct {
	WithdrawID string `json:"withdrawId"`
	Status     string `json:"status"`
}

func (p *WalletService) CancelFiatWithdraw(r *http.Request, request *WalletCancelWithdrawRequest, reply *WalletCancelWithdrawResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "WalletService.CancelFiatWithdraw")
	log = GetServiceRequestLog(log, r, "Wallet", "CancelFiatWithdraw")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = GetSessionCookie(r)
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
	withdrawID, err := sID.FromSecureID("withdraw", sID.Parse(request.WithdrawID))
	if err != nil {
		log.WithError(err).
			WithField("WithdrawID", request.WithdrawID).
			Error("Wrong WithdrawID")
		return sessions.ErrInternalError
	}

	log = log.WithField("WithdrawID", withdrawID)

	wi, err := accounting.UserCancelWithdraw(ctx, uint64(withdrawID))
	if err != nil {
		log.WithError(err).Error("UserCancelWithdraw failed")
		return err
	}

	*reply = WalletCancelWithdrawResponse{
		WithdrawID: request.WithdrawID,
		Status:     wi.Status,
	}

	return nil
}
