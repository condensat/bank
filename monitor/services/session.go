package services

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/networking/sessions"
)

func verifySessionId(ctx context.Context, sessionID sessions.SessionID) (bool, error) {
	log := logger.Logger(ctx).WithField("Method", "verifySessionId")
	nats := appcontext.Messaging(ctx)

	message := bank.ToMessage(appcontext.AppName(ctx), &sessions.SessionInfo{
		SessionID: sessionID,
	})
	if message == nil {
		log.Error("bank.ToMessage failed")
		return false, ErrServiceInternalError
	}

	response, err := nats.Request(ctx, services.VerifySessionSubject, message)
	if err != nil {
		log.WithError(err).
			WithField("Subject", services.VerifySessionSubject).
			Error("nats.Request Failed")
		return false, ErrServiceInternalError
	}

	var si sessions.SessionInfo
	err = bank.FromMessage(response, &si)
	if err != nil {
		log.WithError(err).
			Error("Message data is not SessionInfo")
		return false, ErrServiceInternalError
	}

	return sessionID == si.SessionID &&
		sessions.IsSessionValid(si.SessionID) &&
		sessions.IsUserValid(si.UserID) &&
		!si.Expired(), nil
}
