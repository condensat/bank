package services

import (
	"net/http"

	apiservice "git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"

	"git.condensat.tech/bank/logger"
)

type DashboardService int

// StatusRequest holds args for status requests
type StatusRequest struct {
	apiservice.SessionArgs
}

type UsersStatus struct {
	Count int `json:"count"`
}

// StatusResponse holds args for string requests
type StatusResponse struct {
	Users UsersStatus `json:"users"`
}

func (p *DashboardService) Status(r *http.Request, request *StatusRequest, reply *StatusResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.Status")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "Status")

	userCount, err := database.UserCount(db)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}
	*reply = StatusResponse{
		Users: UsersStatus{
			Count: userCount,
		},
	}

	log.Info("Status")

	return nil
}
