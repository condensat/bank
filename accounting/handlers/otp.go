package handlers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/security/utils"
)

func ValidateOtp(ctx context.Context, authInfo common.AuthInfo) error {
	db := appcontext.Database(ctx)
	if db == nil {
		return errors.New("Invalid Database")
	}

	if len(authInfo.OperatorAccount) == 0 {
		return errors.New("Invalid OperatorAccount")
	}
	if len(authInfo.TOTP) == 0 {
		return errors.New("Invalid TOTP")
	}

	email := fmt.Sprintf("%s@condensat.tech", authInfo.OperatorAccount)

	operator, err := database.FindUserByEmail(db, model.UserEmail(email))
	if err != nil {
		return errors.New("OperatorAccount not found")
	}
	if operator.Name != model.UserName(authInfo.OperatorAccount) {
		return errors.New("Wrong OperatorAccount")
	}

	login := hex.EncodeToString([]byte(utils.HashString(authInfo.OperatorAccount[:])))
	operatorID, valid, err := database.CheckTOTP(ctx, db, model.Base58(login), string(authInfo.TOTP))
	if err != nil {
		return errors.New("CheckTOTP failed")
	}
	if !valid {
		return errors.New("Invalid OTP")
	}
	if operatorID != operator.ID {
		return errors.New("Wrong operator ID")
	}

	return nil
}
