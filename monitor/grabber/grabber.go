package grabber

import (
	"context"
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/monitor"
	"git.condensat.tech/bank/utils"

	"git.condensat.tech/bank/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrAddProcessInfo = errors.New("AddProcessInfo")
	ErrInternalError  = errors.New("InternalError")
)

type Grabber int

func (p *Grabber) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "monitor.Grabber.Run")

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Grabber Service started")

	<-ctx.Done()
}

func (p *Grabber) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "monitor.Grabber.RegisterHandlers")

	messaging := appcontext.Messaging(ctx)
	messaging.SubscribeWorkers(ctx, "Condensat.Monitor.Inbound", 4, p.onProcessInfo)
	log.Debug("Monitor Grabber registered")
}

func (p *Grabber) onProcessInfo(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "monitor.Grabber.onProcessInfo")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var req monitor.ProcessInfo
	err := bank.FromMessage(message, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not ProcessInfo")
		return nil, ErrInternalError
	}

	log = log.WithFields(logrus.Fields{
		"AppName":  req.AppName,
		"Hostname": req.Hostname,
	})

	err = monitor.AddProcessInfo(ctx, &req)
	if err != nil {
		log.WithError(err).Error("Failed to AddProcessInfo")
		return nil, ErrAddProcessInfo
	}

	return nil, nil
}
