package handlers

import (
	"reflect"
	"testing"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
)

func TestFiatValidateWithdraw(t *testing.T) {
	const databaseName = "TestFiatValidateWithdraw"
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

	err := initTestData(db)
	if err != nil {
		return
	}

	// start a withdraw
	_, err = AccountTransferWithdrawFiat(testCtx, common.AccountTransferWithdrawFiat{
		BatchMode: "normal",
		UserID:    uint64(2),
		Source: common.AccountEntry{
			AccountID:        uint64(3),
			Currency:         "CHF",
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
		log.WithError(err).Error("AccountTransferWithdrawFiat failed")
		return
	}

	type args struct {
		ID []uint64
	}
	tests := []struct {
		name    string
		args    args
		want    common.FiatValidWithdrawList
		wantErr bool
	}{
		{"Default", args{}, common.FiatValidWithdrawList{}, true},
		{"Invalid withdrawID", args{[]uint64{100}}, common.FiatValidWithdrawList{}, true},
		{"Valid", args{[]uint64{ValidatedWithdrawCases["Valid"].TargetID}}, ReturnValues["Valid"], false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := FiatValidateWithdraw(testCtx, tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FiatValidateWithdraw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FiatValidateWithdraw() = %+v, want %v", got, tt.want)
			}
		})
	}
}

var ValidatedWithdrawCases = map[string]common.FiatValidWithdraw{
	"Empty": {},
	"Valid": {
		WithdrawID: 1,
		TargetID:   1,
		UserName:   string(customerUser.Name),
		IBAN:       common.IBAN(validIban),
		Amount:     wAmt,
		Currency:   "CHF",
		AccountID:  3,
	},
}

var ReturnValues = map[string]common.FiatValidWithdrawList{
	"Empty": {ValidatedWithdraws: []common.FiatValidWithdraw{
		ValidatedWithdrawCases["Empty"],
	}},
	"Valid": {ValidatedWithdraws: []common.FiatValidWithdraw{
		ValidatedWithdrawCases["Valid"],
	}},
}
