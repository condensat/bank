package bank

import (
	"context"
	"time"
)

type MessageHandler func(ctx context.Context, subject string, message *Message) (*Message, error)

type Messaging interface {
	SubscribeWorkers(ctx context.Context, subject string, workerCount int, handle MessageHandler)
	Subscribe(ctx context.Context, subject string, handle MessageHandler)

	Request(ctx context.Context, subject string, message *Message) (*Message, error)
	RequestWithTimeout(ctx context.Context, subject string, message *Message, timeout time.Duration) (*Message, error)
}
