package client

import (
	"context"

	"git.condensat.tech/bank/accounting/common"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func UserWithdrawsCrypto(ctx context.Context, userID uint64) (common.UserWithdraws, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.UserWithdrawsCrypto")
	log = log.WithField("UserID", userID)

	if userID == 0 {
		return common.UserWithdraws{}, cache.ErrInternalError
	}

	var result common.UserWithdraws
	err := messaging.RequestMessage(ctx, appcontext.AppName(ctx), common.UserWithdrawListSubject, &common.UserWithdraws{UserID: userID}, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.UserWithdraws{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"Count": len(result.Withdraws),
	}).Debug("UserWithdraws request")

	return result, nil
}
