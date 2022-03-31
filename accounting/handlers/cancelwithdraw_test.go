package handlers

import (
	"testing"

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

	err := initTestData(db)
	if err != nil {
		log.WithError(err).Error("initTestData failed")
		return
	}

	// start a withdraw
	_, err = AccountTransferWithdrawCrypto(testCtx, common.AccountTransferWithdrawCrypto{
		BatchMode: "normal",
		Source: common.AccountEntry{
			AccountID:        uint64(4),
			Currency:         "BTC",
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
		log.WithError(err).Error("AccountTransferWithdrawFiat failed")
		return
	}
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

	feeAmount := wAmt * feeRate
	if feeAmount < common.MinAmountFiatWithdraw {
		feeAmount = fiatMinFee
	}

	type args = cancelArgs

	tests := []struct {
		name    string
		args    args
		want    common.WithdrawInfo
		wantErr bool
	}{
		{"Default", cancelCases["Empty"], common.WithdrawInfo{}, true},
		{"Invalid ID", cancelCases["Invalid ID"], common.WithdrawInfo{}, true},
		{"Valid Fiat", cancelCases["Valid Fiat"], common.WithdrawInfo{
			WithdrawID: cancelCases["Valid Fiat"].TargetID,
			Status:     string(model.WithdrawStatusCanceling),
		}, false},
		{"Valid BTC", cancelCases["Valid BTC"], common.WithdrawInfo{
			WithdrawID: cancelCases["Valid BTC"].TargetID,
			Status:     string(model.WithdrawStatusCanceling),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CancelWithdraw(testCtx, uint64(tt.args.TargetID), tt.args.Comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CancelWithdraw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("FiatCancelWithdraw() = %v, want %v", got, tt.want)
			if got.WithdrawID != tt.want.WithdrawID && got.Status != tt.want.Status {
				t.Errorf("CancelWithdraw() = %v, want %v", got, tt.want)
			}
		})
	}
}

type cancelArgs struct {
	TargetID uint64
	Comment  string
}

var cancelCases = map[string]cancelArgs{
	"Empty":      {},
	"Invalid ID": {TargetID: 10, Comment: "test"},
	"Valid BTC":  {TargetID: 1, Comment: "test"},
	"Valid Fiat": {TargetID: 2, Comment: "test"},
}
