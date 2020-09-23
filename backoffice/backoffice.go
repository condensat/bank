package backoffice

import (
	"context"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
)

type BackOffice int

func (p *BackOffice) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "backoffice.BackOffice.Run")

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("BackOffice Service started")

	<-ctx.Done()
}
