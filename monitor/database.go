package monitor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/monitor/common"
	"github.com/jinzhu/gorm"
)

func AddProcessInfo(ctx context.Context, processInfo *common.ProcessInfo) error {
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return errors.New("Wrong database")
	}

	return db.Create(&processInfo).Error
}

func ListServices(ctx context.Context, since time.Duration) ([]string, error) {
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return nil, errors.New("Wrong database")
	}

	var result []string

	var list []*common.ProcessInfo

	now := time.Now().UTC()
	distinctAppName := fmt.Sprintf("distinct (%s)", gorm.ToColumnName("AppName"))
	err := db.Select(distinctAppName).
		Where("timestamp BETWEEN ? AND ?", now.Add(-since), now).
		Find(&list).Error
	if err != nil {
		return nil, err
	}

	for _, entry := range list {
		result = append(result, entry.AppName)
	}

	return result, nil
}
