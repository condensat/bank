package services

import (
	"context"
	"net/http"
	"sort"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/monitor/messaging"

	coreService "git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/api/sessions"

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

// ServiceList operation return the list of active services
func (p *StackService) ServiceList(r *http.Request, request *StackInfoRequest, reply *StackInfoResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.StackService.ServiceList")
	log = coreService.GetServiceRequestLog(log, r, "Stack", "ServiceList")

	verified, err := verifySessionId(ctx, sessions.SessionID(request.SessionID))
	if err != nil {
		log.WithError(err).
			Error("verifySessionId Failed")
		return ErrServiceInternalError
	}

	if !verified {
		log.Error("Invalid sessionId")
		return sessions.ErrInvalidSessionID
	}

	// Request Service List
	listService, err := StackListServiceRequest(ctx)
	if err != nil {
		log.WithError(err).
			Error("StackListRequest Failed")
		return ErrServiceInternalError
	}

	// Reply
	reply.Services = listService.Services[:]

	log.WithFields(logrus.Fields{
		"Services": reply.Services,
	}).Debug("Stack Services")

	return nil
}

func StackListServiceRequest(ctx context.Context) (StackListService, error) {
	log := logger.Logger(ctx).WithField("Method", "services.StackService.StackListServiceRequest")
	nats := appcontext.Messaging(ctx)
	var result StackListService

	message := bank.ToMessage(appcontext.AppName(ctx), &StackListService{
		Since: time.Hour,
	})
	response, err := nats.Request(ctx, messaging.StackListSubject, message)
	if err != nil {
		log.WithError(err).
			WithField("Subject", messaging.StackListSubject).
			Error("nats.Request Failed")
		return result, ErrServiceInternalError
	}

	err = bank.FromMessage(response, &result)
	if err != nil {
		log.WithError(err).
			Error("Message data is not StackListService")
		return result, ErrServiceInternalError
	}

	sort.Slice(result.Services, func(i, j int) bool {
		return result.Services[i] < result.Services[j]
	})

	return result, nil
}
