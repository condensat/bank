package wallet

import (
	"context"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/common"
	"git.condensat.tech/bank/wallet/handlers"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
)

const (
	DefaultInterval time.Duration = 30 * time.Second
)

type Wallet int

func (p *Wallet) Run(ctx context.Context, options WalletOptions) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.Run")

	// add RedisMutext to context
	ctx = cache.RedisMutexContext(ctx)

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Wallet Service started")

	go p.scheduledUpdate(ctx, options.Chains(), DefaultInterval)

	<-ctx.Done()
}

func (p *Wallet) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 4

	nats.SubscribeWorkers(ctx, common.CryptoAddressNextDepositSubject, concurencyLevel, handlers.OnCryptoAddressNextDeposit)

	log.Debug("Bank Wallet registered")
}

func checkParams(interval time.Duration) time.Duration {
	if interval < time.Second {
		interval = DefaultInterval
	}
	return interval
}

func (p *Wallet) scheduledUpdate(ctx context.Context, chains []string, interval time.Duration) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.scheduledUpdate")

	interval = checkParams(interval)

	log = log.WithFields(logrus.Fields{
		"Interval": interval.String(),
	})

	log.Info("Start wallet Scheduler")

	for epoch := range utils.Scheduler(ctx, interval, 0) {
		chainsStates, err := FetchChainsState(ctx, chains...)
		if err != nil {
			log.WithError(err).
				Error("Failed to FetchChainsState")
			continue
		}

		log.WithFields(logrus.Fields{
			"Epoch": epoch.Truncate(time.Millisecond),
			"Count": len(chainsStates),
		}).Info("Chain state fetched")

		err = UpdateRedisChain(ctx, chainsStates)
		if err != nil {
			log.WithError(err).
				Error("Failed to UpdateRedisChain")
			continue
		}
	}
}
