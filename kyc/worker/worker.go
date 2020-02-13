package worker

import (
	"context"
	"errors"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/kyc"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrRequestNotHandled   = errors.New("Request Not Handled")
	ErrWorkerInternalError = errors.New("Worker Internal Error")
)

type Worker int

func (p *Worker) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "kyc.Worker.Run")

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Worker Service started")

	<-ctx.Done()
}

func (p *Worker) RegisterHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "kyc.Worker.RegisterHandlers")

	messaging := appcontext.Messaging(ctx)
	messaging.SubscribeWorkers(ctx, "Kyc.Start", 2, p.NatsHandler)
	log.Debug("Kyc worker registered")
}

func (p *Worker) NatsHandler(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "kyc.Worker.StartKyc")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	switch subject {
	case "Kyc.Start":
		return p.StartKyc(ctx, message)

	default:
		log.Error("Unknown request subject")
		return nil, ErrRequestNotHandled
	}
}

func (p *Worker) StartKyc(ctx context.Context, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "kyc.Worker.StartKyc")

	var req kyc.KycStart
	err := bank.DecodeObject(message.Data, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not KycStart")
		return nil, ErrWorkerInternalError
	}

	log = log.WithFields(logrus.Fields{
		"UserID": req.UserID,
		"Email":  req.Email,
	})

	// Todo - generate kycID
	kycID := uuid.New().String()

	// Todo - Store to database
	resp := kyc.KycStartResponse{
		ID: kycID,
	}

	log.WithField("KycId", kycID).
		Debug("Kyc session started")

	message = bank.ToMessage("Kyc.Worker", &resp)
	if message == nil {
		log.WithError(err).Error("Failed to create response message")
		return nil, ErrWorkerInternalError
	}

	return message, nil
}