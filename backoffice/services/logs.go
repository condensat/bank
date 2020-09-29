package services

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger/model"

	"github.com/jinzhu/gorm"
)

type LogStatus struct {
	Warnings int `json:"warning"`
	Errors   int `json:"errors"`
	Panics   int `json:"panics"`
}

func FetchLogStatus(ctx context.Context) (LogStatus, error) {
	db := appcontext.Database(ctx)

	logsInfo, err := model.LogsInfo(db.DB().(*gorm.DB))

	return LogStatus{
		Warnings: logsInfo.Warnings,
		Errors:   logsInfo.Errors,
		Panics:   logsInfo.Panics,
	}, err
}
