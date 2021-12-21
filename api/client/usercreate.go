package client

import (
	"context"

	"git.condensat.tech/bank/api/common"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/messaging"

	"github.com/sirupsen/logrus"
)

func UserCreate(ctx context.Context, authInfo common.AuthInfo, pgpPublicKey common.PGPPublicKey) (common.UserInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "Client.UserCreate")

	request := common.UserCreation{
		AuthInfo:     authInfo,
		PGPPublicKey: pgpPublicKey,
	}
	var result common.UserCreation
	err := messaging.RequestMessage(ctx, common.UserCreateSubject, &request, &result)
	if err != nil {
		log.WithError(err).
			Error("RequestMessage failed")
		return common.UserInfo{}, messaging.ErrRequestFailed
	}

	log.WithFields(logrus.Fields{
		"AccountNumber": result.UserInfo.AccountNumber,
	}).Debug("User account Created")

	return result.UserInfo, nil
}
