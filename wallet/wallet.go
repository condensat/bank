package wallet

import (
	"context"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/bitcoin"
	"git.condensat.tech/bank/wallet/chain"
	"git.condensat.tech/bank/wallet/common"
	"git.condensat.tech/bank/wallet/handlers"
	"git.condensat.tech/bank/wallet/tasks"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
)

const (
	DefaultChainInterval      time.Duration = 30 * time.Second
	DefaultOperationsInterval time.Duration = 5 * time.Second

	ConfirmedBlockCount   = 3 // number of confirmation to consider transaction complete
	UnconfirmedBlockCount = 6 // number of confirmation to continue fetching addressInfos

	AddressInfoMinConfirmation = 0
	AddressInfoMaxConfirmation = 9999
)

type Wallet int

func (p *Wallet) Run(ctx context.Context, options WalletOptions) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.Run")

	// add RedisMutext to context
	ctx = cache.RedisMutexContext(ctx)

	// load rpc clients configurations
	chainsOptions := loadChainsOptionsFromFile(options.FileName)

	// create all rpc clients
	for _, chainOption := range chainsOptions.Chains {
		log.WithField("Chain", chainOption.Chain).
			Info("Adding rpc client")
		ctx = common.ChainClientContext(ctx, chainOption.Chain, bitcoin.New(ctx, bitcoin.BitcoinOptions{
			ServerOptions: bank.ServerOptions{
				HostName: chainOption.HostName,
				Port:     chainOption.Port,
			},
			User: chainOption.User,
			Pass: chainOption.Pass,
		}))
	}

	p.registerHandlers(ctx)

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
	}).Info("Wallet Service started")

	go mainScheduler(ctx, chainsOptions.Names())

	<-ctx.Done()
}

func (p *Wallet) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	ctx = handlers.ChainHandlerContext(ctx, p)

	const concurencyLevel = 4

	nats.SubscribeWorkers(ctx, common.CryptoAddressNextDepositSubject, concurencyLevel, handlers.OnCryptoAddressNextDeposit)

	log.Debug("Bank Wallet registered")
}

func mainScheduler(ctx context.Context, chains []string) {
	log := logger.Logger(ctx).WithField("Method", "Wallet.mainScheduler")

	taskChainUpdate := utils.Scheduler(ctx, DefaultChainInterval, 0)
	taskOperationsUpdate := utils.Scheduler(ctx, DefaultOperationsInterval, 0)

	for {
		select {

		// update chains
		case epoch := <-taskChainUpdate:
			tasks.UpdateChains(ctx, epoch, chains)

		// update operation
		case epoch := <-taskOperationsUpdate:
			tasks.UpdateOperations(ctx, epoch, chains)

		case <-ctx.Done():
			log.Info("Daemon exited")
			return
		}
	}
}

// common.Chain interface
func (p *Wallet) GetNewAddress(ctx context.Context, chainName, account string) (string, error) {
	return chain.GetNewAddress(ctx, chainName, account)
}
