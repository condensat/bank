package sessions

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func VerifySession(ctx context.Context, subject string, message *messaging.Message) (*messaging.Message, error) {
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
	err = messaging.FromMessage(message, &sessionInfo)
	if err != nil {
		log.WithError(err).
			Warning("Message data is not SessionInfo")
		return nil, ErrInternalError
	}

	resp := session.sessionInfo(ctx, sessionInfo.SessionID)

	return messaging.ToMessage(appcontext.AppName(ctx), &resp), nil
}
