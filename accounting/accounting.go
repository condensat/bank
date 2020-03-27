package accounting

import (
	"context"
	"errors"

	"git.condensat.tech/bank/utils"

	"git.condensat.tech/bank/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrAddProcessInfo = errors.New("AddProcessInfo")
	ErrInternalError  = errors.New("InternalError")
)

type Accounting int

func (p *Accounting) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.Run")

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Accounting Service started")

	<-ctx.Done()
}

func (p *Accounting) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	log.Debug("Bank Accounting registered")
}
