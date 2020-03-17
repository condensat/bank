package sessions

import (
	"context"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"github.com/sirupsen/logrus"
)

func VerifySession(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "services.VerifySession")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	// Retrieve sessions from context
	session, err := ContextSession(ctx)
	if err != nil {
		log.WithError(err).
			Warning("Session renew failed")
		return nil, ErrInternalError
	}

	var sessionInfo SessionInfo
	err = bank.FromMessage(message, &sessionInfo)
	if err != nil {
		log.WithError(err).
			Warning("Message data is not SessionInfo")
		return nil, ErrInternalError
	}

	resp := session.sessionInfo(ctx, sessionInfo.SessionID)

	return bank.ToMessage(appcontext.AppName(ctx), &resp), nil
}
