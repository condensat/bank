package handlers

import (
	"fmt"
	"reflect"
	"testing"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
)

func TestValidateWithdrawTarget(t *testing.T) {
	const databaseName = "TestValidateWithdrawTarget"
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

	fiatAccount, err := database.GetAccountsByUserAndCurrencyAndName(db, 2, "EUR", "default")
	if err != nil {
		log.WithError(err).Error("GetAccountsByUserAndCurrencyAndName failed")
		return
	}

	btcAccount, err := database.GetAccountsByUserAndCurrencyAndName(db, 2, "BTC", "default")
	if err != nil {
		log.WithError(err).Error("GetAccountsByUserAndCurrencyAndName failed")
		return
	}

	err = initWithdraw(testCtx, uint64(btcAccount[0].ID), "BTC")
	if err != nil {
		log.WithError(err).Error("initWithdraw failed for BTC")
		return
	}
	err = initWithdraw(testCtx, uint64(fiatAccount[0].ID), "EUR")
	if err != nil {
		log.WithError(err).Error("initWithdraw failed for EUR")
		return
	}

	feeAmount := wAmt * feeRate
	if feeAmount < common.MinAmountFiatWithdraw {
		feeAmount = fiatMinFee
	}

	var targetFiat model.WithdrawTarget
	var targetBtc model.WithdrawTarget
	// Get the withdraw
	i := 0
	for {
		i += 1
		fmt.Printf("getting withdraw with id %d\n", i)
		withdraw, err := database.GetWithdraw(db, model.WithdrawID(i))
		if err != nil {
			break
		}
		// Get the account
		account, err := database.GetAccountByID(db, withdraw.From)
		if err != nil {
			log.WithError(err).Error("GetAccountByID failed")
			return
		}

		fmt.Printf("account: %+v\n", account)

		// Get the target
		switch account.CurrencyName {
		case "EUR":
			targetFiat, err = database.GetWithdrawTargetByWithdrawID(db, withdraw.ID)
			if err != nil {
				log.WithError(err).Error("GetWithdrawTargetByWithdrawID failed")
				continue
			}
		case "BTC":
			targetBtc, err = database.GetWithdrawTargetByWithdrawID(db, withdraw.ID)
			if err != nil {
				log.WithError(err).Error("GetWithdrawTargetByWithdrawID failed")
				continue
			}
		}

	}

	if targetBtc.ID == 0 || targetFiat.ID == 0 {
		log.Error("Can't get targets")
		return
	}

	type args struct {
		target model.WithdrawTarget
	}
	tests := []struct {
		name    string
		args    args
		want    WithdrawValidationRule
		wantErr bool
	}{
		{"Default", args{}, WithdrawValidationRule{}, true},
		{"fiat break rule", args{targetFiat}, WithdrawValidationRule{
			Amount:         100_000,
			TimeSpan:       Day,
			Action:         Report,
			CurrencyType:   Fiat,
			WithdrawTarget: WithdrawSEPA,
		}, false},
		{"btc break rule", args{targetBtc}, WithdrawValidationRule{
			Amount:         5000,
			TimeSpan:       Day,
			Action:         Report,
			CurrencyType:   CryptoNative,
			WithdrawTarget: WithdrawOnChain,
		}, false},
		// {"Valid", args{targetFiatValid}, WithdrawValidationRule{}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateWithdrawTarget(testCtx, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithdrawTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateWithdrawTarget() = %+v, want %v", got, tt.want)
			}
		})
	}
}
