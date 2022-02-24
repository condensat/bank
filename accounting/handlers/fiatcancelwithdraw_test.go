package handlers

import (
	"fmt"
	"reflect"
	"testing"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
)

func TestFiatCancelWithdraw(t *testing.T) {
	const databaseName = "TestFiatCancelWithdraw"
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

	err := createFiatCancelTestData(db)
	if err != nil {
		log.WithError(err).Error("createFiatCancelTestData failed")
		return
	}

	feeAmount := wAmt * feeRate
	if feeAmount < minAmountFiatWithdraw {
		feeAmount = fiatMinFee
	}

	type args struct {
		fiatCancelWithdraw common.FiatCancelWithdraw
	}
	tests := []struct {
		name    string
		args    args
		want    common.FiatCancelWithdraw
		wantErr bool
	}{
		{"Default", args{cancelCases["Empty"]}, common.FiatCancelWithdraw{}, true},
		{"Invalid ID", args{cancelCases["Invalid ID"]}, common.FiatCancelWithdraw{}, true},
		{"Valid", args{cancelCases["Valid"]}, cancelCases["Valid Return"], false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fiatCancelWithdraw(testCtx, db, log, tt.args.fiatCancelWithdraw)
			if (err != nil) != tt.wantErr {
				t.Errorf("FiatCancelWithdraw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FiatCancelWithdraw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createFiatCancelTestData(db bank.Database) error {
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

		if newAccount.UserID != 1 {
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
			if err != nil {
				return err
			}

			// start a withdraw
			_, err = FiatWithdraw(testCtx, uint64(account.UserID), common.AccountEntry{
				AccountID:        uint64(newAccount.ID),
				Currency:         string(newAccount.CurrencyName),
				Amount:           wAmt,
				OperationType:    string(model.OperationTypeFiatWithdraw),
				SynchroneousType: string(model.SynchroneousTypeSync),
			}, common.FiatSepaInfo{
				IBAN:  common.IBAN(validIban),
				BIC:   common.BIC(validBic),
				Label: "test label",
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var cancelCases = map[string]common.FiatCancelWithdraw{
	"Empty":        {},
	"Invalid ID":   {FiatOperationInfoID: 10, Comment: "test"},
	"Valid":        {FiatOperationInfoID: 1, Comment: "test"},
	"Valid Return": {FiatOperationInfoID: 1, Comment: "test", Amount: wAmt, Currency: "CHF", IBAN: common.IBAN(validIban), UserName: string(customerUser.Name)},
}
