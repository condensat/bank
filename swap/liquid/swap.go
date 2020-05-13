package liquid

import (
	"context"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
)

type Swap int

func (p *Swap) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Swap.Run")

	p.registerHandlers(cache.RedisMutexContext(ctx))

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Liquid Swap Service started")

	<-ctx.Done()
}

func (p *Swap) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.RegisterHandlers")

	log.Debug("Liquid Swap registered")
}
