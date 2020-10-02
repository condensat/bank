package services

import (
	"net/http"

	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/api/sessions"

	apiservice "git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/database/model"
)

// WalletListRequest holds args for walletlist requests
type WalletListRequest struct {
	apiservice.SessionArgs
}

// BatchListResponse holds response for walletlist request
type WalletListResponse struct {
	Wallets []string `json:"wallets"`
}

func (p *DashboardService) WalletList(r *http.Request, request *WalletListRequest, reply *WalletListResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.WalletList")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "WalletList")

	// Get userID from session
	request.SessionID = apiservice.GetSessionCookie(r)
	sessionID := sessions.SessionID(request.SessionID)

	isAdmin, log, err := isUserAdmin(ctx, log, sessionID)
	if err != nil {
		log.WithError(err).
			WithField("RoleName", model.RoleNameAdmin).
			Error("UserHasRole failed")
		return ErrPermissionDenied
	}
	if !isAdmin {
		log.WithError(err).
			Error("User is not Admin")
		return ErrPermissionDenied
	}

	wallets, err := FetchWalletList(ctx)
	if err != nil {
		log.WithError(err).
			Error("FetchWalletList failed")
		return sessions.ErrInternalError
	}

	*reply = WalletListResponse{
		Wallets: wallets,
	}

	return nil
}
