package bank

import (
	"context"
	"time"

	"git.condensat.tech/bank/database/model"
	logModel "git.condensat.tech/bank/logger/model"
)

type Key []byte

type PublicKey Key
type PrivateKey Key
type SharedKey Key

type ServerOptions struct {
	HostName string
	Port     int
}

type Logger interface {
	Close()
	CreateLogEntry(timestamp time.Time, app, level, msg, data string) *logModel.LogEntry
	AddLogEntries(entries []*logModel.LogEntry) error
}

// Messaging (Nats)
type NC interface{}

type MessageHandler func(ctx context.Context, subject string, message *Message) (*Message, error)
type Messaging interface {
	NC() NC

	SubscribeWorkers(ctx context.Context, subject string, workerCount int, handle MessageHandler)
	Subscribe(ctx context.Context, subject string, handle MessageHandler)

	Request(ctx context.Context, subject string, message *Message) (*Message, error)
	RequestWithTimeout(ctx context.Context, subject string, message *Message, timeout time.Duration) (*Message, error)
}

// Database (GORM)
type DB interface{}

type Database interface {
	DB() DB

	Migrate(models []model.Model) error
}

// Cache (Redis)
type RDB interface{}

type Cache interface {
	RDB() RDB
}
