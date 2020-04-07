package accounting

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/accounting/handlers"
	"git.condensat.tech/bank/accounting/internal"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
)

type Accounting int

func (p *Accounting) Run(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.Run")

	p.registerHandlers(internal.RedisMutexContext(ctx))

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Accounting Service started")

	<-ctx.Done()
}

func (p *Accounting) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 8

	nats.SubscribeWorkers(ctx, common.CurrencyInfoSubject, 2*concurencyLevel, handlers.OnCurrencyInfo)
	nats.SubscribeWorkers(ctx, common.CurrencyCreateSubject, 2*concurencyLevel, handlers.OnCurrencyCreate)
	nats.SubscribeWorkers(ctx, common.CurrencyListSubject, 2*concurencyLevel, handlers.OnCurrencyList)
	nats.SubscribeWorkers(ctx, common.CurrencySetAvailableSubject, 2*concurencyLevel, handlers.OnCurrencySetAvailable)

	nats.SubscribeWorkers(ctx, common.AccountCreateSubject, 2*concurencyLevel, handlers.OnAccountCreate)
	nats.SubscribeWorkers(ctx, common.AccountListSubject, 2*concurencyLevel, handlers.OnAccountList)
	nats.SubscribeWorkers(ctx, common.AccountHistorySubject, 2*concurencyLevel, handlers.OnAccountHistory)
	nats.SubscribeWorkers(ctx, common.AccountSetStatusSubject, 2*concurencyLevel, handlers.OnAccountSetStatus)
	nats.SubscribeWorkers(ctx, common.AccountOperationSubject, 8*concurencyLevel, handlers.OnAccountOperation)
	nats.SubscribeWorkers(ctx, common.AccountTransfertSubject, 8*concurencyLevel, handlers.OnAccountTransfert)

	log.Debug("Bank Accounting registered")
}
