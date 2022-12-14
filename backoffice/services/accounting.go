package services

import (
	"context"
	"net/http"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/logger"

	apiservice "git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/api/sessions"

	"code.condensat.tech/bank/secureid"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

const (
	DefaultAccountCountByPage = 10
)

type CurrencyBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Locked   float64 `json:"locked"`
}

type AccountingStatus struct {
	Count    int               `json:"count"`
	Active   int               `json:"active"`
	Balances []CurrencyBalance `json:"balances"`
}

func FetchAccountingStatus(ctx context.Context) (AccountingStatus, error) {
	db := appcontext.Database(ctx)

	accountsInfo, err := database.AccountsInfos(db)
	if err != nil {
		return AccountingStatus{}, err
	}

	var balances []CurrencyBalance
	for _, account := range accountsInfo.Accounts {
		balances = append(balances, CurrencyBalance{
			Currency: account.CurrencyName,
			Balance:  account.Balance,
			Locked:   account.TotalLocked,
		})
	}

	return AccountingStatus{
		Count:    accountsInfo.Count,
		Active:   accountsInfo.Active,
		Balances: balances,
	}, nil
}

// AccountListRequest holds args for accountlist requests
type AccountListRequest struct {
	apiservice.SessionArgs
	RequestPaging
}

type AccountInfo struct {
	AccountID string  `json:"accountId"`
	UserID    string  `json:"userId"`
	Name      string  `json:"name"`
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance"`
	Status    string  `json:"status"`
}

// AccountListResponse holds response for accountlist request
type AccountListResponse struct {
	RequestPaging
	Accounts []AccountInfo `json:"accounts"`
}

func (p *DashboardService) AccountList(r *http.Request, request *AccountListRequest, reply *AccountListResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.AccountList")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "AccountList")

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

	sID := appcontext.SecureID(ctx)

	var startID secureid.Value
	if len(request.Start) > 0 {
		startID, err = sID.FromSecureID("account", sID.Parse(request.Start))
		if err != nil {
			log.WithError(err).
				WithField("Start", request.Start).
				Error("startID FromSecureID failed")
			return ErrPermissionDenied
		}
	}
	var pagesCount int
	var accountPage []model.Account
	infos := make(map[model.AccountID]AccountInfo)
	err = db.Transaction(func(db bank.Database) error {
		var err error
		pagesCount, err = database.AccountPagingCount(db, DefaultAccountCountByPage)
		if err != nil {
			pagesCount = 0
			return err
		}

		accountPage, err = database.AccountPage(db, model.AccountID(startID), DefaultUserCountByPage)
		if err != nil {
			accountPage = nil
			return err
		}
		for _, account := range accountPage {
			var info AccountInfo

			status, err := database.GetAccountStatusByAccountID(db, account.ID)
			if err != nil {
				accountPage = nil
				return err
			}
			last, err := database.GetLastAccountOperation(db, account.ID)
			if err != nil {
				accountPage = nil
				return err
			}

			info.Name = string(account.Name)
			info.Balance = float64(*last.Balance)
			info.Currency = string(account.CurrencyName)
			info.Status = string(status.State)

			infos[account.ID] = info
		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("UserPaging failed")
		return apiservice.ErrServiceInternalError
	}

	var next string
	if len(accountPage) > 0 {
		nextID := int(accountPage[len(accountPage)-1].ID) + 1
		secureID, err := sID.ToSecureID("account", secureid.Value(uint64(nextID)))
		if err != nil {
			return err
		}
		next = sID.ToString(secureID)
	}

	var accounts []AccountInfo
	for _, account := range accountPage {
		// create secureID
		secureID, err := sID.ToSecureID("account", secureid.Value(uint64(account.ID)))
		if err != nil {
			return err
		}
		// create user secureID
		userSecureID, err := sID.ToSecureID("user", secureid.Value(uint64(account.UserID)))
		if err != nil {
			return err
		}

		var accountInfo AccountInfo
		if info, ok := infos[account.ID]; ok {
			accountInfo = info
		}
		accountInfo.AccountID = sID.ToString(secureID)
		accountInfo.UserID = sID.ToString(userSecureID)

		accounts = append(accounts, accountInfo)
	}

	*reply = AccountListResponse{
		RequestPaging: RequestPaging{
			Page:         request.Page,
			PageCount:    pagesCount,
			CountPerPage: DefaultUserCountByPage,
			Start:        request.Start,
			Next:         next,
		},
		Accounts: accounts[:],
	}

	return nil
}

// UserAccountListRequest holds args for useraccountlist requests
type UserAccountListRequest struct {
	apiservice.SessionArgs
	UserID string `json:"userId"`
}

// UserAccountListResponse holds response for useraccountlist request
type UserAccountListResponse struct {
	UserID     string           `json:"userId"`
	Accounts   []string         `json:"accounts"`
	Accounting AccountingStatus `json:"accounting"`
}

func (p *DashboardService) UserAccountList(r *http.Request, request *UserAccountListRequest, reply *UserAccountListResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.UserAccountListRequest")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "UserAccountListRequest")

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

	sID := appcontext.SecureID(ctx)

	userID, err := sID.FromSecureID("user", sID.Parse(request.UserID))
	if err != nil {
		log.WithError(err).
			WithField("UserID", request.UserID).
			Error("userID FromSecureID failed")
		return ErrPermissionDenied
	}

	var user model.User
	var accounts []string
	var accounting AccountingStatus
	err = db.Transaction(func(db bank.Database) error {
		var err error

		user, err = database.FindUserById(db, model.UserID(userID))
		if err != nil {
			return err
		}

		accountsInfo, err := database.AccountsInfosByUser(db, model.UserID(userID))
		if err != nil {
			return err
		}

		var balances []CurrencyBalance
		for _, account := range accountsInfo.Accounts {
			balances = append(balances, CurrencyBalance{
				Currency: account.CurrencyName,
				Balance:  account.Balance,
				Locked:   account.TotalLocked,
			})
		}

		accounting = AccountingStatus{
			Count:    accountsInfo.Count,
			Active:   accountsInfo.Active,
			Balances: balances,
		}

		accountIDs, err := database.GetUserAccounts(db, user.ID)
		if err != nil {
			return err
		}
		for _, accountID := range accountIDs {
			secureID, err := sID.ToSecureID("account", secureid.Value(uint64(accountID)))
			if err != nil {
				return err
			}

			accounts = append(accounts, sID.ToString(secureID))
		}
		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("UserAccountList failed")
		return apiservice.ErrServiceInternalError
	}

	*reply = UserAccountListResponse{
		UserID:     request.UserID,
		Accounts:   accounts,
		Accounting: accounting,
	}

	return nil
}

// AccountDetailRequest holds args for accountdetail requests
type AccountDetailRequest struct {
	apiservice.SessionArgs
	AccountID string `json:"accountId"`
}

// AccountDetailResponse holds response for accountdetail request
type AccountDetailResponse struct {
	AccountID string  `json:"accountId"`
	UserID    string  `json:"userId"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}

func (p *DashboardService) AccountDetail(r *http.Request, request *AccountDetailRequest, reply *AccountDetailResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.UserAccountListRequest")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "UserAccountListRequest")

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

	sID := appcontext.SecureID(ctx)

	accountID, err := sID.FromSecureID("account", sID.Parse(request.AccountID))
	if err != nil {
		log.WithError(err).
			WithField("AccountID", request.AccountID).
			Error("accountID FromSecureID failed")
		return apiservice.ErrServiceInternalError
	}

	var account model.Account
	var accountState model.AccountState
	var last model.AccountOperation
	err = db.Transaction(func(db bank.Database) error {
		var err error

		account, err = database.GetAccountByID(db, model.AccountID(accountID))
		if err != nil {
			return err
		}
		accountState, err = database.GetAccountStatusByAccountID(db, model.AccountID(accountID))
		if err != nil {
			return err
		}

		last, err = database.GetLastAccountOperation(db, model.AccountID(accountID))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.WithError(err).
			Error("AccountDetail failed")
		return apiservice.ErrServiceInternalError
	}

	secureID, err := sID.ToSecureID("user", secureid.Value(uint64(account.UserID)))
	if err != nil {
		return err
	}

	*reply = AccountDetailResponse{
		AccountID: request.AccountID,
		UserID:    sID.ToString(secureID),
		Currency:  string(account.CurrencyName),
		Balance:   float64(*last.Balance),
		Name:      string(account.Name),
		Status:    string(accountState.State),
	}

	return nil
}
