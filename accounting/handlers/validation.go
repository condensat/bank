package handlers

import (
	"context"
	"errors"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/currency/rate"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"github.com/sirupsen/logrus"
)

var (
	ErrUnconsistentCurrency       = errors.New("Unconsistent currency")
	ErrNotImplementedCurrencyType = errors.New("Not Implemented currency type")
)

const (
	ValidationBaseCurrency = "CHF"
)

type ValidationAmount float64

type ValidationTimeSpan uint64

const (
	Day ValidationTimeSpan = iota
	Month
	Year
)

func (vts ValidationTimeSpan) String() string {
	switch vts {
	case Day:
		return "Day"
	case Month:
		return "Month"
	case Year:
		return "Year"
	default:
		return "Unknown time frame"
	}
}

func (vts ValidationTimeSpan) GetStartTime() time.Time {
	now := time.Now().UTC()

	switch vts {
	case Day:
		return time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	case Month:
		return time.Date(now.Year(), now.Month()-1, now.Day(), 0, 0, 0, 0, time.UTC)
	case Year:
		return time.Date(now.Year()-1, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	default:
		return time.Date(0, time.Month(1), 0, 0, 0, 0, 0, time.UTC)
	}
}

type ValidationCurrency uint64

// We have to keep this identical to Currency Type in database/model/currency.go
const (
	Fiat ValidationCurrency = iota
	CryptoNative
	CryptoAsset
)

func (vc ValidationCurrency) String() string {
	switch vc {
	case Fiat:
		return "Fiat"
	case CryptoNative:
		return "Crypto Native"
	case CryptoAsset:
		return "Crypto Asset"
	default:
		return "Unknown currency type"
	}
}

type ValidationWithdrawTarget uint64

// See database/model/withdrawtarget.go
const (
	WithdrawSEPA ValidationWithdrawTarget = iota
	WithdrawSwift
	WithdrawCard
	WithdrawOnChain
	WithdrawLiquid
	WithdrawLightning
)

func (vwt ValidationWithdrawTarget) String() string {
	switch vwt {
	case WithdrawSEPA:
		return string(model.WithdrawTargetSepa)
	case WithdrawSwift:
		return string(model.WithdrawTargetSwift)
	case WithdrawCard:
		return string(model.WithdrawTargetCard)
	case WithdrawOnChain:
		return string(model.WithdrawTargetOnChain)
	case WithdrawLiquid:
		return string(model.WithdrawTargetLiquid)
	case WithdrawLightning:
		return string(model.WithdrawTargetLightning)
	default:
		return "Unknown Withdraw Target"
	}
}

type ValidationAction uint64

const (
	Ask ValidationAction = iota
	Report
	Multisig
)

func (va ValidationAction) String() string {
	switch va {
	case Ask:
		return "Ask"
	case Report:
		return "Report"
	case Multisig:
		return "Multisig"
	default:
		return "Unknown Action"
	}
}

type WithdrawValidationRule struct {
	Amount         ValidationAmount
	TimeSpan       ValidationTimeSpan
	CurrencyType   ValidationCurrency
	WithdrawTarget ValidationWithdrawTarget
	Action         ValidationAction
}

func BitcoinDailyWithdrawLimit() WithdrawValidationRule {
	return WithdrawValidationRule{
		Amount:         5_000,
		TimeSpan:       Day,
		Action:         Report,
		CurrencyType:   CryptoNative,
		WithdrawTarget: WithdrawOnChain,
	}
}

func BitcoinYearWithdrawLimit() WithdrawValidationRule {
	return WithdrawValidationRule{
		Amount:         100_000,
		TimeSpan:       Year,
		Action:         Report,
		CurrencyType:   CryptoNative,
		WithdrawTarget: WithdrawOnChain,
	}
}

func SepaDailyWithdrawLimit() WithdrawValidationRule {
	return WithdrawValidationRule{
		Amount:         100_000,
		TimeSpan:       Day,
		Action:         Report,
		CurrencyType:   Fiat,
		WithdrawTarget: WithdrawSEPA,
	}
}

func SepaYearWithdrawLimit() WithdrawValidationRule {
	return WithdrawValidationRule{
		Amount:         100_000,
		TimeSpan:       Year,
		Action:         Report,
		CurrencyType:   Fiat,
		WithdrawTarget: WithdrawSEPA,
	}
}

func getWithdrawRules() []WithdrawValidationRule {
	var result []WithdrawValidationRule

	result = append(result, BitcoinDailyWithdrawLimit())
	result = append(result, BitcoinYearWithdrawLimit())
	result = append(result, SepaDailyWithdrawLimit())
	result = append(result, SepaYearWithdrawLimit())

	return result
}

func ValidateWithdrawTarget(ctx context.Context, target model.WithdrawTarget) (WithdrawValidationRule, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.ValidateWithdrawTarget")

	db := appcontext.Database(ctx)

	var result WithdrawValidationRule

	if target.WithdrawID == 0 {
		err := errors.New("target is empty")
		log.Error(err)
		return result, err
	}

	// get the withdraw
	withdraw, err := database.GetWithdraw(db, target.WithdrawID)
	if err != nil {
		log.WithError(err).
			Error("GetWithdraw failed")
		return result, err
	}

	// get the account
	account, err := database.GetAccountByID(db, model.AccountID(withdraw.From))
	if err != nil {
		log.WithError(err).
			Error("GetAccountByID failed")
		return result, err
	}

	// Get the userID
	user, err := database.FindUserById(db, account.UserID)
	if err != nil {
		log.WithError(err).Error("FindUserByID failed")
		return result, err
	}

	// Now check against the known rules
	rules := getWithdrawRules()

	for _, rule := range rules {

		if rule.WithdrawTarget.String() != string(target.Type) {
			continue
		}

		start := rule.TimeSpan.GetStartTime()

		// Get the validation entries for Withdraw operations from start time
		validations, err := database.GetWithdrawValidationsFromStartToNow(db, user.ID, start, model.WithdrawTargetType(rule.WithdrawTarget.String()))
		if err != nil {
			log.WithError(err).Error("GetWithdrawValidationsFromStartToNow failed")
			return result, err
		}

		var totalAmount float64
		// Get the current operation amount in base currency
		rateBase, err := rate.GetLatestRateForBase(ctx, string(account.CurrencyName), ValidationBaseCurrency)
		if err != nil {
			log.WithError(err).Error("GetLatestRateForBase failed")
			return result, err
		}
		precision, err := database.GetCurrencyPrecision(db, account.CurrencyName)
		if err != nil {
			log.WithError(err).Error("GetCurrencyPrecision failed")
			return result, err
		}

		baseAmount := rate.ConvertWithRate(float64(*withdraw.Amount), rateBase, precision)
		if err != nil {
			log.WithError(err).Error("ConvertWithRate failed")
			return result, err
		}

		totalAmount += baseAmount

		// compare the sum of all withdraws against the rule
		for _, v := range validations {
			totalAmount += float64(*v.Amount)
		}

		if totalAmount > float64(rule.Amount) {
			log.WithFields(logrus.Fields{
				"withdrawID":    withdraw.ID,
				"Rule Timespan": rule.TimeSpan,
				"Rule Target":   rule.WithdrawTarget,
				"Rule Amount":   rule.Amount,
				"Total amount":  totalAmount,
			}).Warnf("Withdraw needs manual validation")
			return rule, nil
		}
	}

	return result, nil
}
