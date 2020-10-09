package services

import (
	"context"

	"git.condensat.tech/bank/appcontext"

	"git.condensat.tech/bank/api/sessions"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"github.com/sirupsen/logrus"
)

func isUserAdmin(ctx context.Context, log *logrus.Entry, sessionID sessions.SessionID) (bool, *logrus.Entry, error) {

	db := appcontext.Database(ctx)
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return false, log, err
	}

	// Get userID from session
	userID := session.UserSession(ctx, sessionID)
	if !sessions.IsUserValid(userID) {
		log.Error("Invalid userSession")
		return false, log, sessions.ErrInvalidSessionID
	}

	log = log.WithFields(logrus.Fields{
		"SessionID": sessionID,
		"UserID":    userID,
	})

	isAdmin, err := database.UserHasRole(db, model.UserID(userID), model.RoleNameAdmin)
	if err != nil {
		log.WithError(err).
			WithField("RoleName", model.RoleNameAdmin).
			Error("UserHasRole failed")
		return false, log, ErrPermissionDenied
	}

	return isAdmin, log, nil
}
