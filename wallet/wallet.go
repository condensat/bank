package wallet

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/common"
	"git.condensat.tech/bank/wallet/handlers"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
)

type Wallet int

func (p *Wallet) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.Run")

	p.registerHandlers(cache.RedisMutexContext(ctx))

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Wallet Service started")

	<-ctx.Done()
}

func (p *Wallet) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 4

	nats.SubscribeWorkers(ctx, common.CryptoAddressNextDepositSubject, concurencyLevel, handlers.OnCryptoAddressNextDeposit)

	log.Debug("Bank Wallet registered")
}
