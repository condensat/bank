package logger

import (
	"time"

	"git.condensat.tech/bank/logger/model"
)

type Logger interface {
	Close()
	CreateLogEntry(timestamp time.Time, app, level string, userID uint64, sessionID string, method, err, msg, data string) *model.LogEntry
	AddLogEntries(entries []*model.LogEntry) error
}
