package handlers

import (
	"math"
	"testing"
	"time"

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

	// Now we already make a deposit here, because it's more convenient for other tests, maybe that makes this test kind of pointless maybe?
	err := initTestData(db)
	if err != nil {
		log.WithError(err).Error("initTestData failed")
		return
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
			if (err != nil) != tt.wantErr {
				t.Errorf("FiatDeposit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !func(got, want common.AccountEntry) bool {
				if got.Amount != want.Amount ||
					got.Balance != want.Balance ||
					got.LockAmount != want.LockAmount ||
					got.TotalLocked != want.TotalLocked ||
					got.AccountID != want.AccountID ||
					got.Currency != want.Currency ||
					got.SynchroneousType != want.SynchroneousType ||
					got.OperationType != want.OperationType ||
					got.Label != want.Label {
					return false
				}
				return true
			}(got, tt.want) {
				t.Errorf("FiatDeposit() = %v, want %v", got, tt.want)
			}
		})
	}
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
	OperationID:      uint64((len(testUsers)*len(currencies))*2 + 1),
	AccountID:        uint64(len(currencies) + 1),
	Currency:         string(currencies[0].Name),
	ReferenceID:      2,
	OperationType:    string(model.OperationTypeFiatDeposit),
	SynchroneousType: "sync",
	Timestamp:        common.Timestamp(),
	Label:            "N/A",
	Amount:           initAmount,
	Balance:          initAmount * 2,
}
