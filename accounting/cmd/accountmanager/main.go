// simply push log entry to redis
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/monitor/processus"

	"git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/accounting/common"

	"github.com/sirupsen/logrus"
)

const (
	SpecimenIban        = "CH5604835012345678009"
	SpecimenInvalidIban = "CH56XXX"
	SpecimenBic         = "KBAGCH22XXX"
	SpecimenUser        = "8868029921"
)

type Args struct {
	App appcontext.Options

	Redis cache.RedisOptions
	Nats  messaging.NatsOptions
}

func parseArgs() Args {
	var args Args
	appcontext.OptionArgs(&args.App, "AccountManager")

	cache.OptionArgs(&args.Redis)
	messaging.OptionArgs(&args.Nats)

	flag.Parse()

	return args
}

func exists(limit int, predicate func(i int) bool) bool {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return true
		}
	}
	return false
}

func createAndListAccount(ctx context.Context, currencies []common.CurrencyInfo, userID uint64) {
	log := logger.Logger(ctx).WithField("Method", "createAndListAccount")

	log = log.WithField("UserID", userID)

	// list user accounts
	userAccounts, err := client.AccountList(ctx, userID)
	if err != nil {
		log.WithError(err).
			Error("ListAccounts Failed")
		return
	}

	accounts := userAccounts.Accounts
	// create account for available currencies
	for _, currency := range currencies {
		if !currency.Available {
			continue
		}

		if exists(len(accounts), func(i int) bool {
			return accounts[i].Currency.Name == currency.Name
		}) {
			continue
		}
		account, err := client.AccountCreate(ctx, userID, currency.Name)
		if err != nil {
			log.WithError(err).
				Error("CreateAccount Failed")
			continue
		}
		log.WithField("Account", fmt.Sprintf("%+v", account)).
			Info("Account Created")

		// change account status to normal
		_, err = client.AccountSetStatus(ctx, account.Info.AccountID, "normal")
		if err != nil {
			log.WithError(err).
				Error("AccountSetStatus Failed")
			continue
		}

		_, err = client.AccountDepositSync(ctx, account.Info.AccountID, 42, 10.0, "First Deposit")
		if err != nil {
			log.WithError(err).
				Error("AccountDeposit Failed")
			continue
		}
	}

	for _, account := range accounts {

		// force write
		for i := 0; i < 1; i++ {
			client.AccountDepositSync(ctx, account.AccountID, 42, 0.1, "Batch Deposit")
			client.AccountDepositSync(ctx, account.AccountID, 42, -0.1, "Batch Deposit")
		}

		if account.AccountID > 4 {
			_, err = client.AccountTransfer(ctx, account.AccountID, 1+(account.AccountID-1)%4, 1337, account.Currency.Name, 0.01, "For weedcoder")
			if err != nil {
				log.WithError(err).
					Error("AccountTransfer Failed")
			}
		}

		to := time.Now()
		from := to.Add(-time.Hour)
		history, err := client.AccountHistory(ctx, account.AccountID, from, to)
		if err != nil {
			log.WithError(err).
				Error("AccountHistory Failed")
			return
		}

		log.WithFields(logrus.Fields{
			"AccountID": account.AccountID,
			"Count":     len(history.Entries),
		}).Infof("Account history")
	}
}

func CreateAccounts(ctx context.Context) {

	// list all currencies
	list, err := client.CurrencyList(ctx)
	if err != nil {
		panic(err)
	}

	var count int

	const userCount = 100
	for userID := 1; userID <= userCount; userID++ {
		createAndListAccount(ctx, list.Currencies, uint64(userID))
	}
	if userCount > 0 {
		return
	}

	start := time.Now()
	for i := 0; i < 10; i++ {
		// create users
		users := make([]uint64, 0, userCount)
		for userID := 1; userID <= userCount; userID++ {
			users = append(users, uint64(userID))
		}
		// randomize
		for i := len(users) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			users[i], users[j] = users[j], users[i]
		}
		batchSize := 128
		batches := make([][]uint64, 0, (len(users)+batchSize-1)/batchSize)
		for batchSize < len(users) {
			users, batches = users[batchSize:], append(batches, users[0:batchSize:batchSize])
		}
		batches = append(batches, users)

		for _, userIDs := range batches {

			var wait sync.WaitGroup
			for _, userID := range userIDs {
				wait.Add(1)

				go func(userID uint64) {
					defer wait.Done()

					createAndListAccount(ctx, list.Currencies, userID)
				}(uint64(userID))

				count++
			}
			wait.Wait()
		}
	}

	fmt.Printf("%d calls in %s\n", count, time.Since(start).Truncate(time.Millisecond))
}

func AccountTransferWithdraw(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "AccountTransferWithdraw")

	const accountID uint64 = 18
	log.WithField("AccountID", accountID)
	withdrawID, err := client.AccountTransferWithdrawCrypto(ctx,
		accountID, "TBTC", 0.00000300, "normal", "Test AccountTransferWithdraw",
		"bitcoin-testnet", "tb1qqjv0dec9vagycgwpchdkxsnapl9uy92dek4nau",
	)
	if err != nil {
		log.WithError(err).
			Error("AccountTransferWithdraw Failed")
		return
	}

	log.
		WithField("withdrawID", withdrawID).
		Info("AccountTransferWithdraw")
}

func DepositFiat(ctx context.Context) {

	userName := SpecimenUser
	amount := 50.0
	currency := "CHF"
	label := "label"

	deposit, err := client.FiatDeposit(ctx, common.AuthInfo{}, userName, amount, currency, label)
	if err != nil {
		fmt.Printf("DepositFiat failed with error: %v\n", err)
		return
	}

	fmt.Printf("Successfully deposited %v %s for user %s\n", deposit.Amount, deposit.Currency, userName)
}

func WithdrawFiat(ctx context.Context) {

	userName := SpecimenUser
	amount := 50.0
	currency := "CHF"
	label := "label"
	iban := SpecimenIban
	// iban := SpecimenInvalidIban
	bic := SpecimenBic
	userLabel := "label"

	withdraw, err := client.FiatWithdraw(ctx, common.AuthInfo{}, userName, amount, currency, label, iban, bic, userLabel)
	if err != nil {
		fmt.Printf("WithdrawFiat failed with error: %v\n", err)
		return
	}
	fmt.Printf("Successfully withdrawed %v %s for user %s\n", withdraw.Amount, withdraw.Currency, userName)
}

func FetchPendingWithdraw(ctx context.Context) {

	withdraws, err := client.FiatFetchPendingWithdraw(ctx, common.AuthInfo{})
	if err != nil {
		fmt.Printf("FetchPendingWithdraw failed with error: %v\n", err)
		return
	}

	if len(withdraws.PendingWithdraws) == 0 {
		fmt.Printf("There's no pending withdraws for now\n")
		return
	}

	for i, withdraw := range withdraws.PendingWithdraws {
		fmt.Printf("\n\nWithdraw #%v: ", i)
		fmt.Printf("\nUserName: %v", withdraw.UserName)
		fmt.Printf("\nIBAN: %v", withdraw.IBAN)
		fmt.Printf("\nBIC: %v", withdraw.BIC)
		fmt.Printf("\nCurrency: %v", withdraw.Currency)
		fmt.Printf("\nAmount: %v", withdraw.Amount)
	}

	fmt.Println()

}

func FinalizeWithdraw(ctx context.Context) {
	userName := SpecimenUser
	iban := SpecimenIban

	final, err := client.FiatFinalizeWithdraw(ctx, common.AuthInfo{}, userName, iban)
	if err != nil {
		fmt.Printf("FinalizeWithdraw failed with error: %v\n", err)
		return
	}

	fmt.Printf("Successfully finalized %v %s withdraw for user %s\n", final.Amount, final.Currency, userName)
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithCache(ctx, cache.NewRedis(ctx, args.Redis))
	ctx = appcontext.WithWriter(ctx, logger.NewRedisLogger(ctx))
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, args.Nats))
	ctx = appcontext.WithProcessusGrabber(ctx, processus.NewGrabber(ctx, 15*time.Second))

	// CreateAccounts(ctx)
	// AccountTransferWithdraw(ctx)
	DepositFiat(ctx)
	WithdrawFiat(ctx)
	FetchPendingWithdraw(ctx)
	FinalizeWithdraw(ctx)
}
