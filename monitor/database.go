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

	now := time.Now().UTC()
	distinctAppName := fmt.Sprintf("distinct (%s)", gorm.ToColumnName("AppName"))

	var list []*common.ProcessInfo
	err := db.Select(distinctAppName).
		Where("timestamp BETWEEN ? AND ?", now.Add(-since), now).
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var result []string
	for _, entry := range list {
		result = append(result, entry.AppName)
	}

	return result, nil
}

func LastServicesStatus(ctx context.Context) ([]common.ProcessInfo, error) {
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return nil, errors.New("Wrong database")
	}

	subQuery := db.Model(&common.ProcessInfo{}).
		Select("MAX(id) as id, MAX(timestamp) AS last").
		Where("timestamp >= DATE_SUB(NOW(), INTERVAL 3 MINUTE)").
		Group("app_name, hostname").
		SubQuery()

	var list []*common.ProcessInfo
	err := db.Joins("RIGHT JOIN (?) AS t1 ON process_info.id = t1.id AND timestamp = t1.last", subQuery).
		Order("app_name ASC, hostname ASC, timestamp DESC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var result []common.ProcessInfo
	for _, entry := range list {
		result = append(result, *entry)
	}

	return result, nil
}

func LastServiceHistory(ctx context.Context, appName string, from, to time.Time, step time.Duration, round time.Duration) ([]common.ProcessInfo, error) {
	db, ok := appcontext.Database(ctx).DB().(*gorm.DB)
	if !ok {
		return nil, errors.New("Wrong database")
	}

	tsFrom := from.UnixNano() / int64(time.Second)
	tsTo := to.UnixNano() / int64(time.Second)

	subQuery := db.Model(&common.ProcessInfo{}).
		Select("MAX(id) AS id, FLOOR(UNIX_TIMESTAMP(timestamp)/(?)) AS timekey", step/time.Second).
		Where("app_name=?", appName).
		Where("timestamp BETWEEN FROM_UNIXTIME(?) AND FROM_UNIXTIME(?)", tsFrom, tsTo).
		Group("timekey, hostname").
		SubQuery()

	var list []*common.ProcessInfo
	err := db.Joins("RIGHT JOIN (?) AS t1 ON process_info.id = t1.id", subQuery).
		Order("timestamp, hostname DESC").
		Find(&list).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var result []common.ProcessInfo
	for _, entry := range list {
		entry.Timestamp = entry.Timestamp.Round(round)
		result = append(result, *entry)
	}

	return result, nil
}
