package api

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"
	"git.condensat.tech/bank/logger"

	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidUserInfo        = errors.New("Invalid user info")
	ErrInvalidLoginOrPassword = errors.New("Invalid login or password")
	ErrInvalidEmail           = errors.New("Invalid email")
)

type UserInfo struct {
	Login,
	Password,
	Email string
	Roles []string
}

func ParseUserInfo(userInfo string) (UserInfo, error) {
	toks := strings.Split(userInfo, ":")
	if len(toks) != 4 {
		return UserInfo{}, ErrInvalidUserInfo
	}

	login := toks[0]
	password := toks[1]
	if len(login) == 0 || len(password) == 0 {
		return UserInfo{}, ErrInvalidLoginOrPassword
	}

	email := toks[2]
	_, err := mail.ParseAddress(fmt.Sprintf("%s <%s>", login, email))
	if err != nil {
		return UserInfo{}, ErrInvalidEmail
	}

	roles := strings.Split(toks[3], ",")
	if len(roles) == 0 {
		roles = append(roles, "user")
	}

	return UserInfo{
		Login:    login,
		Password: password,
		Email:    email,
		Roles:    roles,
	}, nil
}

func scannerFromFileOrStdin(fileName string) (*bufio.Scanner, *os.File, error) {
	if len(fileName) == 0 || fileName == "-" {
		return bufio.NewScanner(os.Stdin), nil, nil
	} else {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, nil, err
		}
		return bufio.NewScanner(file), file, nil
	}
}

func FromUserInfoFile(ctx context.Context, fileName string) ([]UserInfo, error) {
	log := logger.Logger(ctx).WithField("Method", "api.FromUserInfoFile")
	scanner, file, err := scannerFromFileOrStdin(fileName)
	if err != nil {
		return nil, err
	}
	if file != nil {
		defer file.Close()
	}

	var result []UserInfo
	for scanner.Scan() {
		userInfo, err := ParseUserInfo(scanner.Text())
		if err != nil {
			log.WithError(err).
				Error("Failed to ParseUserInfo")
			continue
		}
		result = append(result, userInfo)

		if userInfo.Login == "demo" {
			for i := 0; i < 100; i++ {
				demo := userInfo
				demo.Login = fmt.Sprintf("%s_%.3d", demo.Login, i)
				demo.Email = fmt.Sprintf("%s@condensat.space", demo.Login)
				result = append(result, demo)
			}
		}
	}
	return result[:], nil
}

func ImportUsers(ctx context.Context, userInfos ...UserInfo) error {
	log := logger.Logger(ctx).WithField("Method", "api.ImportUsers")
	db := appcontext.Database(ctx)
	if db == nil {
		return errors.New("Invalid Database")
	}

	return db.Transaction(func(tx bank.Database) error {

		batchSize := 32
		batches := make([][]UserInfo, 0, (len(userInfos)+batchSize-1)/batchSize)
		for batchSize < len(userInfos) {
			userInfos, batches = userInfos[batchSize:], append(batches, userInfos[0:batchSize:batchSize])
		}
		batches = append(batches, userInfos)

		for _, userInfos := range batches {
			for _, userInfo := range userInfos {
				user, err := database.FindOrCreateUser(tx, model.User{
					Name:  model.UserName(userInfo.Login),
					Email: model.UserEmail(userInfo.Email),
				})
				if err != nil {
					log.WithError(err).
						Error("Failed to FindOrCreateUser")
					continue
				}

				credential, err := database.CreateOrUpdatedCredential(ctx, tx,
					model.Credential{
						UserID:       user.ID,
						LoginHash:    model.Base58(userInfo.Login),
						PasswordHash: model.Base58(userInfo.Password),
						TOTPSecret:   "",
					},
				)
				if err != nil {
					log.WithError(err).
						Error("Failed to CreateOrUpdatedCredential")
					continue
				}

				userID, verified, err := database.CheckCredential(ctx, tx,
					database.HashEntry(model.Base58(userInfo.Login)),
					database.HashEntry(model.Base58(userInfo.Password)),
				)
				if err != nil {
					log.WithError(err).
						Error("Failed to CheckCredential")
					continue
				}

				if !verified {
					log.Error("Not Verified")
					continue
				}

				if userID != user.ID {
					log.Error("Wrong UserID")
					continue
				}

				log.WithFields(logrus.Fields{
					"UserID":       userID,
					"LoginHash":    credential.LoginHash,
					"PasswordHash": credential.PasswordHash,
					"Verified":     verified,
				}).Info("User Imported")
			}
		}
		return nil
	})
}
