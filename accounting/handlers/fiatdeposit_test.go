package handlers

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
)

func TestFiatDeposit(t *testing.T) {
	const databaseName = "TestFiatDeposit"
	t.Parallel()

	db := setup(databaseName, database.WithdrawModel())
	defer teardown(db, databaseName)

	log := logger.Logger(testCtx)

	testCtx = common.BankUserContext(testCtx, bankUser)
	testCtx = appcontext.WithDatabase(testCtx, db)
	redisOptions := cache.RedisOptions{}
	cache.OptionArgs(&redisOptions)
	testCtx = appcontext.WithCache(testCtx, cache.NewRedis(testCtx, redisOptions))
	testCtx = cache.RedisMutexContext(testCtx)

	err := createFiatDepositTestData(db)
	if err != nil {
		log.WithError(err).Error("createFiatDepositTestData failed")
		return
	}

	feeAmount := wAmt * feeRate
	if feeAmount < minAmountFiatWithdraw {
		feeAmount = fiatMinFee
	}

	type args struct {
		deposit common.FiatDeposit
	}
	tests := []struct {
		name    string
		args    args
		want    common.AccountEntry
		wantErr bool
	}{
		{"Default", args{depositCases["Empty"]}, common.AccountEntry{}, true},
		{"Invalid Username", args{depositCases["Invalid Username"]}, common.AccountEntry{}, true},
		{"Negative Amount", args{depositCases["Negative Amount"]}, common.AccountEntry{}, true},
		{"Absurd Amount", args{depositCases["Absurd Amount"]}, common.AccountEntry{}, true},
		{"Non Fiat Currency", args{depositCases["Non Fiat Currency"]}, common.AccountEntry{}, true},
		{"Invalid OperationType", args{depositCases["Invalid OperationType"]}, common.AccountEntry{}, true},
		{"Valid", args{depositCases["Valid"]}, validReturn, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FiatDeposit(testCtx, tt.args.deposit.UserName, tt.args.deposit.Destination)
			fmt.Printf("err: %s\n", err)
			if (err != nil) != tt.wantErr {
				t.Errorf("FiatDeposit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FiatDeposit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createFiatDepositTestData(db bank.Database) error {
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
	}

	return nil
}

var depositCases = map[string]common.FiatDeposit{
	"Empty": {},
	"Invalid Username": {UserName: "Invalid", Destination: common.AccountEntry{
		OperationType:    string(model.OperationTypeFiatDeposit),
		SynchroneousType: string(model.SynchroneousTypeSync),
		Timestamp:        time.Now().UTC().Truncate(time.Second),
		Label:            "test",
		Amount:           100.0,
		LockAmount:       0.0,
		Currency:         string(currencies[0].Name),
	}},
	"Negative Amount": {UserName: string(customerUser.Name), Destination: common.AccountEntry{
		OperationType:    string(model.OperationTypeFiatDeposit),
		SynchroneousType: string(model.SynchroneousTypeSync),
		Timestamp:        time.Now().UTC().Truncate(time.Second),
		Label:            "test",
		Amount:           -initAmount,
		LockAmount:       0.0,
		Currency:         string(currencies[0].Name),
	}},
	"Absurd Amount": {UserName: string(customerUser.Name), Destination: common.AccountEntry{
		OperationType:    string(model.OperationTypeFiatDeposit),
		SynchroneousType: string(model.SynchroneousTypeSync),
		Timestamp:        time.Now().UTC().Truncate(time.Second),
		Label:            "test",
		Amount:           math.MaxFloat64,
		LockAmount:       0.0,
		Currency:         string(currencies[0].Name),
	}},
	"Non Fiat Currency": {UserName: string(customerUser.Name), Destination: common.AccountEntry{
		OperationType:    string(model.OperationTypeFiatDeposit),
		SynchroneousType: string(model.SynchroneousTypeSync),
		Timestamp:        time.Now().UTC().Truncate(time.Second),
		Label:            "test",
		Amount:           initAmount,
		LockAmount:       0.0,
		Currency:         string(currencies[1].Name),
	}},
	"Invalid OperationType": {UserName: string(customerUser.Name), Destination: common.AccountEntry{
		OperationType:    string(model.OperationTypeTransfer),
		SynchroneousType: string(model.SynchroneousTypeSync),
		Timestamp:        time.Now().UTC().Truncate(time.Second),
		Label:            "test",
		Amount:           initAmount,
		LockAmount:       0.0,
		Currency:         string(currencies[0].Name),
	}},
	"Valid": {UserName: string(customerUser.Name), Destination: common.AccountEntry{
		OperationType:    string(model.OperationTypeFiatDeposit),
		SynchroneousType: string(model.SynchroneousTypeSync),
		Timestamp:        time.Now().UTC().Truncate(time.Second),
		Label:            "test",
		Amount:           initAmount,
		LockAmount:       0.0,
		Currency:         string(currencies[0].Name),
	}},
}

var validReturn common.AccountEntry = common.AccountEntry{
	OperationID:      uint64(len(testUsers)*len(currencies) + 1),
	AccountID:        uint64(len(currencies) + 1),
	Currency:         string(currencies[0].Name),
	ReferenceID:      2,
	OperationType:    string(model.OperationTypeFiatDeposit),
	SynchroneousType: "sync",
	Timestamp:        common.Timestamp(),
	Label:            "N/A",
	Amount:           initAmount,
	Balance:          initAmount,
}
