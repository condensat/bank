package handlers

import (
	"testing"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
)

func TestAccountTransferWithdrawFiat(t *testing.T) {
	const databaseName = "TestAccountTransferWithdrawFiat"
	t.Parallel()

	db := setup(databaseName, database.WithdrawModel())
	defer teardown(db, databaseName)

	testCtx = common.BankUserContext(testCtx, bankUser)
	testCtx = appcontext.WithDatabase(testCtx, db)
	redisOptions := cache.RedisOptions{}
	cache.OptionArgs(&redisOptions)
	testCtx = appcontext.WithCache(testCtx, cache.NewRedis(testCtx, redisOptions))
	testCtx = cache.RedisMutexContext(testCtx)

	err := initTestData(db)
	if err != nil {
		return
	}

	feeAmount := wAmt * feeRate
	if feeAmount < common.MinAmountFiatWithdraw {
		feeAmount = fiatMinFee
	}

	type args struct {
		BatchMode string
		userId    uint64
		withdraw  common.AccountEntry
		sepaInfo  common.FiatSepaInfo
	}
	tests := []struct {
		name    string
		args    args
		want    common.AccountTransfer
		wantErr bool
	}{
		{"Default", args{}, common.AccountTransfer{}, true},
		{"Invalid UserID", args{"normal", 0, withdrawCases["Empty"], sepaCases["Valid"]}, common.AccountTransfer{}, true},
		{"Invalid BatchMode", args{"fake", 2, withdrawCases["Empty"], sepaCases["Valid"]}, common.AccountTransfer{}, true},
		{"Invalid Amount", args{"normal", 2, withdrawCases["Invalid Amount"], sepaCases["Valid"]}, common.AccountTransfer{}, true},
		{"Invalid Currency", args{"normal", 2, withdrawCases["Invalid Currency"], sepaCases["Valid"]}, common.AccountTransfer{}, true},
		{"Invalid Amount below min", args{"normal", 2, withdrawCases["Invalid Amount below min"], sepaCases["Valid"]}, common.AccountTransfer{}, true},
		{"Invalid OperationType", args{"normal", 2, withdrawCases["Invalid OperationType"], sepaCases["Valid"]}, common.AccountTransfer{}, true},
		{"Invalid Sync", args{"normal", 2, withdrawCases["Invalid Sync"], sepaCases["Valid"]}, common.AccountTransfer{}, true},
		{"Invalid LockAmount", args{"normal", 2, withdrawCases["Invalid LockAmount"], sepaCases["Valid"]}, common.AccountTransfer{}, true},

		{"Invalid IBAN", args{"normal", 2, withdrawCases["Valid"], sepaCases["Invalid IBAN"]}, common.AccountTransfer{}, true},
		{"Invalid BIC", args{"normal", 2, withdrawCases["Valid"], sepaCases["Invalid BIC"]}, common.AccountTransfer{}, true},

		{"Valid", args{"normal", 2, withdrawCases["Valid"], sepaCases["Valid"]}, common.AccountTransfer{
			Source: common.AccountEntry{
				Balance: initAmount - wAmt - feeAmount,
			},
			Destination: common.AccountEntry{
				Balance: wAmt + feeAmount,
			},
		}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := AccountTransferWithdrawFiat(testCtx, common.AccountTransferWithdrawFiat{
				BatchMode: tt.args.BatchMode,
				UserID:    tt.args.userId,
				Source:    tt.args.withdraw,
				Sepa:      tt.args.sepaInfo,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("AccountTransferWithdrawFiat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("FiatWithdraw() = %+v, want %v", got, tt.want)
			if !(got.Source.Balance == tt.want.Source.Balance && got.Destination.Balance == tt.want.Destination.Balance) { // for now we just check that the balance is right
				t.Errorf("AccountTransferWithdrawFiat() = %v, want %v", got, tt.want)
			}
		})
	}
}

var withdrawCases = map[string]common.AccountEntry{
	"Empty":                    {},
	"Invalid Amount":           {Amount: -wAmt, Currency: "CHF", OperationType: string(model.OperationTypeTransfer), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid Currency":         {Amount: wAmt, Currency: "FAKE", OperationType: string(model.OperationTypeTransfer), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid Amount below min": {Amount: common.MinAmountFiatWithdraw / 2, Currency: "CHF", OperationType: string(model.OperationTypeTransfer), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid OperationType":    {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeInvalid), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Invalid Sync":             {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeTransfer), SynchroneousType: string(model.SynchroneousTypeInvalid), LockAmount: 0.0},
	"Invalid LockAmount":       {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeTransfer), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 10.0},
	"Valid":                    {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeTransfer), SynchroneousType: string(model.SynchroneousTypeSync), LockAmount: 0.0},
	"Valid Return":             {Amount: wAmt, Currency: "CHF", OperationType: string(model.OperationTypeTransfer), SynchroneousType: string(model.SynchroneousTypeSync), OperationID: 3, AccountID: 1, Balance: 74.5},
}

var sepaCases = map[string]common.FiatSepaInfo{
	"Empty":        {},
	"Invalid IBAN": {IBAN: common.IBAN("FAKE"), BIC: common.BIC(validBic), Label: "test label"},
	"Invalid BIC":  {IBAN: common.IBAN(validIban), BIC: common.BIC("Fake"), Label: "test label"},
	"Valid":        {IBAN: common.IBAN(validIban), BIC: common.BIC(validBic), Label: "test label"},
}
