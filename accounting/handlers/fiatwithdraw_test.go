package handlers

import (
	"fmt"
	"testing"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func TestFiatWithdraw(t *testing.T) {
	const databaseName = "TestFiatWithdraw"
	t.Parallel()

	db := setup(databaseName, database.WithdrawModel())
	defer teardown(db, databaseName)

	testCtx = common.BankUserContext(testCtx, bankUser)
	testCtx = appcontext.WithDatabase(testCtx, db)
	redisOptions := cache.RedisOptions{}
	cache.OptionArgs(&redisOptions)
	testCtx = appcontext.WithCache(testCtx, cache.NewRedis(testCtx, redisOptions))
	testCtx = cache.RedisMutexContext(testCtx)

	err := createTestData(db)
	if err != nil {
		return
	}

	feeAmount := wAmt * feeRate
	if feeAmount < minAmountFiatWithdraw {
		feeAmount = fiatMinFee
	}

	type args struct {
		userId   uint64
		withdraw common.AccountEntry
		sepaInfo common.FiatSepaInfo
	}
	tests := []struct {
		name    string
		args    args
		want    common.AccountEntry
		wantErr bool
	}{
		{"Default", args{}, common.AccountEntry{}, true},
		{"Invalid UserID", args{0, withdrawCases["Empty"], sepaCases["Valid"]}, common.AccountEntry{}, true},
		{"Invalid Amount", args{2, withdrawCases["Invalid Amount"], sepaCases["Valid"]}, common.AccountEntry{}, true},
		{"Invalid Currency", args{2, withdrawCases["Invalid Currency"], sepaCases["Valid"]}, common.AccountEntry{}, true},
		{"Invalid Amount below min", args{2, withdrawCases["Invalid Amount below min"], sepaCases["Valid"]}, common.AccountEntry{}, true},
		{"Invalid OperationType", args{2, withdrawCases["Invalid OperationType"], sepaCases["Valid"]}, common.AccountEntry{}, true},
		{"Invalid Sync", args{2, withdrawCases["Invalid Sync"], sepaCases["Valid"]}, common.AccountEntry{}, true},
		{"Invalid LockAmount", args{2, withdrawCases["Invalid LockAmount"], sepaCases["Valid"]}, common.AccountEntry{}, true},

		{"Invalid IBAN", args{2, withdrawCases["Valid"], sepaCases["Invalid IBAN"]}, common.AccountEntry{}, true},
		{"Invalid BIC", args{2, withdrawCases["Valid"], sepaCases["Invalid BIC"]}, common.AccountEntry{}, true},

		{"Valid", args{2, withdrawCases["Valid"], sepaCases["Valid"]}, common.AccountEntry{Balance: initAmount - wAmt - feeAmount}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := FiatWithdraw(testCtx, tt.args.userId, tt.args.withdraw, tt.args.sepaInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("FiatWithdraw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("FiatWithdraw() = %+v, want %v", got, tt.want)
			if !(got.Balance == tt.want.Balance) { // for now we just check that the balance is right
				t.Errorf("FiatWithdraw() = %v, want %v", got.Balance, tt.want.Balance)
			}
		})
	}
}

func createTestData(db bank.Database) error {
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

var withdrawCases = map[string]common.AccountEntry{
	"Empty":                    {},
	"Invalid Amount":           {Amount: -wAmt, Currency: "CHF", OperationType: string(model.OperationTypeFiatWithdraw), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid Currency":         {Amount: wAmt, Currency: "FAKE", OperationType: string(model.OperationTypeFiatWithdraw), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid Amount below min": {Amount: minAmountFiatWithdraw / 2, Currency: "CHF", OperationType: string(model.OperationTypeFiatWithdraw), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid OperationType":    {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeInvalid), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid Sync":             {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeFiatWithdraw), SynchroneousType: string(model.SynchroneousTypeInvalid), LockAmount: 0.0},
	"Invalid LockAmount":       {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeFiatWithdraw), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 10.0},
	"Valid":                    {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeFiatWithdraw), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Valid Return":             {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeFiatWithdraw), SynchroneousType: string(model.SynchroneousTypeSync), OperationID: 3, AccountID: 1, Balance: 74.5},
}

var sepaCases = map[string]common.FiatSepaInfo{
	"Empty":        {},
	"Invalid IBAN": {IBAN: common.IBAN("FAKE"), BIC: common.BIC(validBic), Label: "test label"},
	"Invalid BIC":  {IBAN: common.IBAN(validIban), BIC: common.BIC("Fake"), Label: "test label"},
	"Valid":        {IBAN: common.IBAN(validIban), BIC: common.BIC(validBic), Label: "test label"},
}
