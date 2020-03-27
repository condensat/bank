package accounting

import (
	"context"
	"time"
)

func ListUserAccounts(ctx context.Context, userID uint64) ([]AccountInfo, error) {
	var result []AccountInfo
	return result, nil
}

func GetAccountHistory(ctx context.Context, accountID uint64, from, to time.Time) ([]AccountEntry, error) {
	var result []AccountEntry
	return result, nil
}
