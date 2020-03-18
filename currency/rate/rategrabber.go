package rate

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
)

type RateGrabber struct {
	appID string
}

func (p *RateGrabber) Run(ctx context.Context, appID string) {
	log := logger.Logger(ctx).WithField("Method", "currency.rate.RateGrabber.Run")
	p.appID = appcontext.SecretOrPassword(appID)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("RateGrabber started")

	<-ctx.Done()
}
