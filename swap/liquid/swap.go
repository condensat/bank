package liquid

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"git.condensat.tech/bank/swap/liquid/common"
	"git.condensat.tech/bank/swap/liquid/handlers"

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

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 8

	nats.SubscribeWorkers(ctx, common.SwapCreateProposalSubject, 2*concurencyLevel, handlers.OnCreateSwapProposal)
	nats.SubscribeWorkers(ctx, common.SwapInfoProposalSubject, 2*concurencyLevel, handlers.OnInfoSwapProposal)

	log.Debug("Liquid Swap registered")
}
