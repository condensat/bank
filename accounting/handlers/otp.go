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

func ValidateOtp(ctx context.Context, authInfo common.AuthInfo, command common.Command) (uint64, error) {
	var result uint64

	db := appcontext.Database(ctx)
	if db == nil {
		return result, errors.New("Invalid Database")
	}

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

	// Register authentication action in operator table
	operatorEntry, err := database.AddOperator(db, model.Operator{
		UserID:    operator.ID,
		Timestamp: common.Timestamp(),
		Command:   model.String(command.String()),
	})
	if err != nil {
		return result, err
	}

	result = uint64(operatorEntry.ID)

	return result, nil
}

func UpdateOperatorTable(ctx context.Context, operatorID, accountID, accountOperationID uint64) error {
	db := appcontext.Database(ctx)
	if db == nil {
		return errors.New("Invalid Database")
	}

	_, err := database.UpdateOperator(db, model.Operator{
		ID:                 model.OperatorID(operatorID),
		AccountID:          model.AccountID(accountID),
		AccountOperationID: model.AccountOperationID(accountOperationID),
	})
	if err != nil {
		return err
	}

	return nil
}
