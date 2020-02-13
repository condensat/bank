package services

import (
	"context"
	"errors"
	"net/http"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/api/sessions"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/kyc"
	"git.condensat.tech/bank/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidKycEmail = errors.New("InvalidKycEmail")
)

// KYCService receiver
type KycService int

// KycStartRequest holds args for start requests
type KycStartRequest struct {
	SessionArgs
	Email string `json:"email"`
}

// KycStartResponse holds args for start requests
type KycStartResponse struct {
	KycID string `json:"kycId"`
}

// Close operation close the session and set status to closed
func (p *KycService) Start(r *http.Request, request *KycStartRequest, reply *KycStartResponse) error {
	ctx := r.Context()
	log := logger.Logger(ctx).WithField("Method", "services.KycService.Start")
	log = GetServiceRequestLog(log, r, "Kyc", "Start")

	if len(request.Email) == 0 {
		log.WithError(ErrInvalidKycEmail).
			Error("Can not start KYC")
		return ErrInvalidKycEmail
	}
	// Retrieve context values
	_, session, err := ContextValues(ctx)
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

	// Request KycID from email
	kycID, err := SendKycIdRequest(ctx, userID, request.Email)
	if err != nil {
		log.WithError(err).
			Error("SendKycIdRequest Failed")
		return ErrServiceInternalError
	}

	// Reply
	*reply = KycStartResponse{
		KycID: kycID,
	}

	log.WithFields(logrus.Fields{
		"KycID": reply.KycID,
	}).Info("Kyc started")

	return nil
}

func SendKycIdRequest(ctx context.Context, userId uint64, email string) (string, error) {
	messaging := appcontext.Messaging(ctx)

	request := bank.ToMessage("Bank.Api", &kyc.KycStart{
		UserID: userId,
		Email:  email,
	})
	message, err := messaging.Request(ctx, "Kyc.Start", request)
	if err != nil {
		return "", err
	}
	if len(message.Error) > 0 {
		return "", errors.New(message.Error)
	}

	var response kyc.KycStartResponse
	err = response.Decode(message.Data)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}
