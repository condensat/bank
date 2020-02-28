package grabber

import (
	"context"
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/monitor"
	"git.condensat.tech/bank/monitor/common"
	"git.condensat.tech/bank/monitor/messaging"
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

	nats := appcontext.Messaging(ctx)
	nats.SubscribeWorkers(ctx, messaging.InboundSubject, 4, p.onProcessInfo)
	nats.SubscribeWorkers(ctx, messaging.StackListSubject, 4, p.onStackList)

	log.Debug("Monitor Grabber registered")
}

func (p *Grabber) onProcessInfo(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "monitor.Grabber.onProcessInfo")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var req common.ProcessInfo
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

func (p *Grabber) onStackList(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "monitor.Grabber.onStackList")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var req common.StackListService
	err := bank.FromMessage(message, &req)
	if err != nil {
		log.WithError(err).Error("Message data is not StackListService")
		return nil, ErrInternalError
	}

	list, err := monitor.ListServices(ctx, req.Since)
	if err != nil {
		log.WithError(err).Error("ListServices failed")
		return nil, ErrInternalError
	}

	resp := common.StackListService{
		Services: list,
	}

	return bank.ToMessage(appcontext.AppName(ctx), &resp), nil
}
