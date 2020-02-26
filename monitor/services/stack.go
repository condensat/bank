package services

import (
	"net/http"

	"git.condensat.tech/bank/api/sessions"
	"git.condensat.tech/bank/logger"

	coreService "git.condensat.tech/bank/api/services"

	"github.com/sirupsen/logrus"
)

// StackService receiver
type StackService int

// StackInfoRequest holds args for start requests
type StackInfoRequest struct {
	coreService.SessionArgs
}

// StackInfoResponse holds args for start requests
type StackInfoResponse struct {
	Services []string `json:"services"`
}

// Info operation return user's email
func (p *StackService) ServiceList(r *http.Request, request *StackInfoRequest, reply *StackInfoResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.StackService.ServiceList")
	log = coreService.GetServiceRequestLog(log, r, "User", "Info")

	// Retrieve context values
	_, session, err := coreService.ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
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

	// Request Service List
	// Todo - Call messaging "Condensat.Monitor.Stack.List"

	// Reply
	*reply = StackInfoResponse{
		Services: []string{"foo", "bar"},
	}

	log.WithFields(logrus.Fields{
		"Services": reply.Services,
	}).Debug("Stack Services")

	return nil
}
