package services

import (
	"net/http"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/database/query"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/networking"
	"git.condensat.tech/bank/networking/sessions"

	"github.com/sirupsen/logrus"
)

// KYCService receiver
type UserService int

// UserInfoRequest holds args for start requests
type UserInfoRequest struct {
	sessions.SessionArgs
}

// UserInfoResponse holds args for start requests
type UserInfoResponse struct {
	Email string `json:"email"`
}

// Info operation return user's email
func (p *UserService) Info(r *http.Request, request *UserInfoRequest, reply *UserInfoResponse) error {
	ctx := r.Context()
	db := appcontext.Database(ctx)
	log := logger.Logger(ctx).WithField("Method", "services.UserService.Info")
	log = networking.GetServiceRequestLog(log, r, "User", "Info")

	// Retrieve context values
	_, session, err := ContextValues(ctx)
	if err != nil {
		log.WithError(err).
			Error("ContextValues Failed")
		return ErrServiceInternalError
	}

	// Get userID from session
	request.SessionID = sessions.GetSessionCookie(r)
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

	// Request UserID from email
	user, err := query.FindUserById(db, model.UserID(userID))
	if err != nil {
		log.WithError(err).
			Error("database.FindUserById Failed")
		return ErrServiceInternalError
	}

	// Reply
	*reply = UserInfoResponse{
		Email: string(user.Email),
	}

	log.WithFields(logrus.Fields{
		"Email": reply.Email,
	}).Info("User started")

	return nil
}
