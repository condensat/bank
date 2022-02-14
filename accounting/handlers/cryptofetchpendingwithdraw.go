package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"
	"git.condensat.tech/bank/security/utils"
	"github.com/sirupsen/logrus"
)

func CryptoFetchPendingWithdraw(ctx context.Context, authInfo common.AuthInfo) ([]common.CryptoWithdraw, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.FetchPendingWithdraws")

	var result []common.CryptoWithdraw

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

	if common.WithOperatorAuth {
		if len(authInfo.OperatorAccount) == 0 {
			return result, errors.New("Invalid OperatorAccount")
		}
		if len(authInfo.TOTP) == 0 {
			return result, errors.New("Invalid TOTP")
		}

		email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

		operator, err := database.FindUserByEmail(db, model.UserEmail(email))
		if err != nil {
			return result, errors.New("OperatorAccount not found")
		}
		if operator.Name != model.UserName(authInfo.OperatorAccount) {
			return result, errors.New("Wrong OperatorAccount")
		}

		login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
		operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
		if err != nil {
			return result, errors.New("CheckTOTP failed")
		}
		if !valid {
			return result, errors.New("Invalid OTP")
		}
		if operatorID != operator.ID {
			return result, errors.New("Wrong operator ID")
		}
	}

	// Fetch the withdraws target
	wt, err := database.GetLastWithdrawTargetByStatus(db, model.WithdrawStatusCreated)
	if err != nil {
		return result, err
	}

	// with withdraws ID, we can fetch Withdraws
	for _, withdraw := range wt {
		// get withdraw
		w, err := database.GetWithdraw(db, withdraw.WithdrawID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetWithdraw")
			return result, err
		}
		// Get withdraw info history
		history, err := database.GetWithdrawHistory(db, withdraw.WithdrawID)
		if err != nil {
			log.WithError(err).
				Error("Failed to GetWithdrawHistory")
			return result, errors.New("error")
		}
		// skip processed withdraw
		if len(history) != 1 || history[0].Status != model.WithdrawStatusCreated {
			log.Warn("Withdraw status is not created")
			continue
		}

		// get data
		data, err := withdraw.OnChainData()
		if err != nil {
			log.WithError(err).
				Error("Failed to get OnChainData")
			return result, errors.New("error")
		}

		// Get userName
		accountID := w.From

		accountInfo, err := database.GetAccountByID(db, accountID)
		if err != nil {
			return result, err
		}

		userInfo, err := database.FindUserById(db, accountInfo.UserID)

		userName := userInfo.Name

		result = append(result, common.CryptoWithdraw{
			WithdrawID: uint64(withdraw.WithdrawID),
			TargetID:   uint64(withdraw.ID),
			UserName:   string(userName),
			Address:    data.PublicKey,
			Amount:     float64(*w.Amount),
			Currency:   string(accountInfo.CurrencyName),
		})
	}

	return result, nil
}

func OnCryptoFetchPendingWithdraw(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Accounting.OnCryptoFetchPendingWithdraw")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.AuthInfo
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			list, err := CryptoFetchPendingWithdraw(ctx, request)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to CryptoFetchPendingWithdraw")
				return nil, cache.ErrInternalError
			}

			log.Info("CryptoFetchPendingWithdraw succeeded")

			// create & return response
			return &common.CryptoFetchPendingWithdrawList{
				PendingWithdraws: list[:],
			}, nil
		})
}
