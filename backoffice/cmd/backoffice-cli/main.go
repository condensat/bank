package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/backoffice"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/messaging"

	"git.condensat.tech/bank/logger"
	logmodel "git.condensat.tech/bank/logger/model"

	"github.com/jinzhu/gorm"
)

type Args struct {
	App appcontext.Options

	Redis    cache.RedisOptions
	Nats     messaging.NatsOptions
	Database database.Options
}

func parseArgs() Args {
	var args Args

	appcontext.OptionArgs(&args.App, "BackOfficeCli")

	database.OptionArgs(&args.Database)

	flag.Parse()

	return args
}

func main() {
	args := parseArgs()

	ctx := context.Background()
	ctx = appcontext.WithOptions(ctx, args.App)
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(args.Database))

	migrateDatabase(ctx)

	AccountsInfo(ctx)
	UsersInfo(ctx)
	DepositList(ctx)
	BatchList(ctx)
	WithdrawList(ctx)
	SwapList(ctx)
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)

	err := db.Migrate(backoffice.Models())
	if err != nil {
		logger.Logger(ctx).WithError(err).
			WithField("Method", "main.migrateDatabase").
			Panic("Failed to migrate backoffice models")
	}
}

func AccountsInfo(ctx context.Context) {
	db := appcontext.Database(ctx)

	gdb := db.DB().(*gorm.DB)
	logsInfo, err := logmodel.LogsInfo(gdb)
	if err != nil {
		panic(err)
	}
	fmt.Printf("LogsInfo: %+v\n", logsInfo)

	userCount, err := database.UserCount(db)
	if err != nil {
		panic(err)
	}
	fmt.Println("UserCount", userCount)

	accountsInfo, err := database.AccountsInfos(db)
	if err != nil {
		panic(err)
	}
	for _, account := range accountsInfo.Accounts {
		fmt.Printf("Accounts: %+v\n", account)
	}
	fmt.Printf("\tCount: %d\n", accountsInfo.Count)
	fmt.Printf("\tActive: %d\n", accountsInfo.Active)

	batchsInfo, err := database.BatchsInfos(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Batchs: %+v\n", batchsInfo)

	withdrawsInfo, err := database.WithdrawsInfos(db)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Withdraws: %+v\n", withdrawsInfo)
}

func UsersInfo(ctx context.Context) {
	db := appcontext.Database(ctx)

	pages, err := database.UserPagingCount(db, 5)
	if err != nil {
		panic(err)
	}
	fmt.Printf("User Pages: %d\n", pages)

	var start string
	for page := 0; page < pages; page++ {
		startID, _ := strconv.Atoi(start)
		userPage, err := database.UserPage(db, model.UserID(startID), 5)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Page %d: Users %+v\n", page, userPage)
		if len(userPage) == 0 {
			break
		}
		start = fmt.Sprintf("%d", int(userPage[len(userPage)-1].ID)+1)
	}
}

func DepositList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := database.DepositPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Deposit Page: %d\n", page)
}

func BatchList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := database.BatchPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Batch Page: %d\n", page)
}

func WithdrawList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := database.WithdrawPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Withdraw Page: %d\n", page)
}

func SwapList(ctx context.Context) {
	db := appcontext.Database(ctx)

	page, err := database.SwapPage(db, 0, 10)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Swap Page: %d\n", page)
}
