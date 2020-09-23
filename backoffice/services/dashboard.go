package services

import (
	"net/http"

	apiservice "git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/api/sessions"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"github.com/sirupsen/logrus"

	"git.condensat.tech/bank/logger"
)

type DashboardService int

// StatusRequest holds args for status requests
type StatusRequest struct {
	apiservice.SessionArgs
}

type UsersStatus struct {
	Count     int `json:"count"`
	Connected int `json:"connected"`
}

// StatusResponse holds args for string requests
type StatusResponse struct {
	Users UsersStatus `json:"users"`
}

func (p *DashboardService) Status(r *http.Request, request *StatusRequest, reply *StatusResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.DashboardService.Status")
	log = apiservice.GetServiceRequestLog(log, r, "Dashboard", "Status")

	db := appcontext.Database(ctx)
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}

	userCount, err := database.UserCount(db)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}
	sessionCount, err := session.Count(ctx)
	if err != nil {
		return apiservice.ErrServiceInternalError
	}

	*reply = StatusResponse{
		Users: UsersStatus{
			Count:     userCount,
			Connected: sessionCount,
		},
	}

	log.WithFields(logrus.Fields{
		"UserCount":    userCount,
		"SessionCount": sessionCount,
	}).Info("Status")

	return nil
}
