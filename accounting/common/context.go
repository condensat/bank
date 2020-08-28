package common

import (
	"context"

	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/database/model"
)

const (
	BankUserKey = "Accounting.BankUser"
)

func BankUserContext(ctx context.Context, bankUser model.User) context.Context {
	return context.WithValue(ctx, BankUserKey, &bankUser)
}

func BankUserFromContext(ctx context.Context) model.User {
	switch bankUser := ctx.Value(BankUserKey).(type) {
	case *model.User:
		return *bankUser

	default:
		logger.Logger(ctx).
			Panic("Unable to get BankUser from context")

		return model.User{}
	}
}
