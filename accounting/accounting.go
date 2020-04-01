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

	nats.SubscribeWorkers(ctx, common.CurrencyCreateSubject, 8, handlers.OnCurrencyCreate)
	nats.SubscribeWorkers(ctx, common.CurrencyListSubject, 8, handlers.OnCurrencyList)
	nats.SubscribeWorkers(ctx, common.CurrencySetAvailableSubject, 8, handlers.OnCurrencySetAvailable)

	nats.SubscribeWorkers(ctx, common.AccountCreateSubject, 8, handlers.OnAccountCreate)
	nats.SubscribeWorkers(ctx, common.AccountListSubject, 8, handlers.OnAccountList)
	nats.SubscribeWorkers(ctx, common.AccountHistorySubject, 8, handlers.OnAccountHistory)
	nats.SubscribeWorkers(ctx, common.AccountSetStatusSubject, 8, handlers.OnAccountSetStatus)
	nats.SubscribeWorkers(ctx, common.AccountOperationSubject, 16, handlers.OnAccountOperation)
	nats.SubscribeWorkers(ctx, common.AccountTransfertSubject, 16, handlers.OnAccountTransfert)

	log.Debug("Bank Accounting registered")
}
