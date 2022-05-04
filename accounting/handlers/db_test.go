package handlers

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sort"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"github.com/jinzhu/gorm"
)

var testCtx = context.Background()
var testUsers = []model.User{
	bankUser,
	customerUser,
}
var bankUser = model.User{
	ID:    model.UserID(1),
	Name:  "CondensatBank",
	Email: "bank@condensat.tech",
}
var customerUser = model.User{
	ID:    model.UserID(2),
	Name:  "12345678901",
	Email: "12345678901@condensat.tech",
}

var validIban = "CH5604835012345678009"
var validBic = "BCNNCH22"

var currencies = []model.Currency{
	model.NewCurrency("CHF", "swiss franc", 0, 1, 0, 2),
	model.NewCurrency("BTC", "bitcoin", 1, 1, 1, 8),
	model.NewCurrency("EUR", "euro", 0, 1, 0, 2),
}

const (
	initAmount = 100_500.0
	wAmt       = 100_001.0
	feeRate    = 0.002
	fiatMinFee = 0.5
)

func setup(databaseName string, models []model.Model) bank.Database {
	options := database.Options{
		HostName:      "mariadb",
		Port:          3306,
		User:          "condensat",
		Password:      "condensat",
		Database:      "condensat",
		EnableLogging: false,
	}
	if databaseName == options.Database {
		panic("Wrong databaseName")
	}

	db := database.NewDatabase(options)
	gdb := db.DB().(*gorm.DB)

	createDatabase := fmt.Sprintf("create database if not exists %s; use %s;", databaseName, databaseName)
	gdb.Exec(createDatabase)

	err := gdb.Exec(createDatabase).Error
	if err != nil {
		panic(err)
	}

	migrateDatabase(db, models)

	return db
}

func teardown(db bank.Database, databaseName string) {
	gdb := db.DB().(*gorm.DB)

	dropDatabase := fmt.Sprintf("drop database if exists %s", databaseName)
	err := gdb.Exec(dropDatabase).Error
	if err != nil {
		panic(err)
	}
}

func migrateDatabase(db bank.Database, models []model.Model) {
	err := db.Migrate(models)
	if err != nil {
		panic(err)
	}
}

func getSortedTypeFileds(t reflect.Type) []string {
	count := t.NumField()
	result := make([]string, 0, count)

	for i := 0; i < count; i++ {
		field := gorm.TheNamingStrategy.Column(t.Field(i).Name)
		result = append(result, field)
	}

	for i, field := range result {
		result[i] = gorm.TheNamingStrategy.Column(field)
	}
	sort.Strings(result)

	return result
}

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func initTestData(db bank.Database) error {
	for _, currency := range currencies {
		_, err := database.AddOrUpdateCurrency(db, currency)
		if err != nil {
			fmt.Println("Can't create currency in db")
			return err
		}
		_, err = database.AddOrUpdateFeeInfo(db, model.FeeInfo{
			Currency: currency.Name,
			Minimum:  fiatMinFee,
			Rate:     feeRate,
		})
		if err != nil {
			fmt.Println("Can't create feeInfo in db")
			return err
		}
	}

	users := []model.User{
		bankUser,
		customerUser,
	}

	var accounts []model.Account
	for _, user := range users {
		newUser, err := database.FindOrCreateUser(db, user)
		if err != nil {
			fmt.Println("Can't create user in db")
			return err
		}
		for _, currency := range currencies {
			accounts = append(accounts, model.Account{
				UserID:       newUser.ID,
				CurrencyName: currency.Name,
				Name:         "default",
			})
		}
	}

	for _, account := range accounts {
		newAccount, err := database.CreateAccount(db, account)
		if err != nil {
			fmt.Println("Can't create account in db")
			return err
		}

		_, err = database.AddOrUpdateAccountState(db, model.AccountState{
			AccountID: newAccount.ID,
			State:     model.AccountStatusNormal,
		})
		if err != nil {
			fmt.Println("Can't set account state")
			return err
		}

		// Fund the new account
		_, err = AccountOperation(testCtx, common.AccountEntry{
			OperationType:    string(model.OperationTypeFiatDeposit),
			SynchroneousType: string(model.SynchroneousTypeSync),

			Label: "init_deposit",

			Amount:     initAmount,
			LockAmount: 0.0,
			Currency:   string(account.CurrencyName),
			AccountID:  uint64(newAccount.ID),
		})

	}

	return nil
}

func initWithdraw(ctx context.Context, accountID uint64, currency string) error {
	db := appcontext.Database(ctx)

	var curr model.Currency
	var err error
	// get currency from db
	if database.CurrencyExists(db, model.CurrencyName(currency)) {
		curr, err = database.GetCurrencyByName(db, model.CurrencyName(currency))
		if err != nil {
			return err
		}
	} else {
		return errors.New("Currency doesn't exist")
	}

	if *curr.Crypto == 1 {
		_, err := AccountTransferWithdrawCrypto(ctx, common.AccountTransferWithdrawCrypto{
			BatchMode: "normal",
			Source: common.AccountEntry{
				AccountID:        accountID,
				Currency:         currency,
				Amount:           wAmt,
				OperationType:    string(model.OperationTypeTransfer),
				SynchroneousType: string(model.SynchroneousTypeSync),
			},
			Crypto: common.CryptoTransfert{
				Chain:     "bitcoin",
				PublicKey: "bc1qxatfze4d2ahf692xhspa42gaaachyq3sf9gaku",
			},
		})
		if err != nil {
			return err
		}
	} else {
		_, err := AccountTransferWithdrawFiat(ctx, common.AccountTransferWithdrawFiat{
			BatchMode: "normal",
			UserID:    uint64(2),
			Source: common.AccountEntry{
				AccountID:        accountID,
				Currency:         currency,
				Amount:           wAmt,
				OperationType:    string(model.OperationTypeTransfer),
				SynchroneousType: string(model.SynchroneousTypeSync),
			},
			Sepa: common.FiatSepaInfo{
				IBAN:  common.IBAN(validIban),
				BIC:   common.BIC(validBic),
				Label: "test label",
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
