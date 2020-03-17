package logger

import (
	"context"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger/model"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type DatabaseLogger struct {
	database bank.Database
	db       *gorm.DB
}

func NewDatabaseLogger(ctx context.Context) *DatabaseLogger {
	database := appcontext.Database(ctx)
	db, ok := database.DB().(*gorm.DB)
	if !ok {
		log.
			Panic("Database is not gorm")
	}

	err := model.Migrate(db)
	if err != nil {
		log.
			WithError(err).
			Panic("Failed to migrate database")
	}

	ret := DatabaseLogger{
		database: database,
		db:       db,
	}

	return &ret
}

func (p *DatabaseLogger) Close() {
	p.db.Close()
}

func (p *DatabaseLogger) CreateLogEntry(timestamp time.Time, app, level string, userID uint64, sessionID string, method, err, msg, data string) *model.LogEntry {
	return &model.LogEntry{
		Timestamp: timestamp.UTC().Round(time.Second),
		App:       app,
		Level:     level,
		UserID:    userID,
		SessionID: sessionID,
		Method:    method,
		Error:     err,
		Message:   msg,
		Data:      data,
	}
}

func (p *DatabaseLogger) AddLogEntries(entries []*model.LogEntry) error {
	return model.TxAddLogEntries(p.db, entries)
}
