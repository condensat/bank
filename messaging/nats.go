package messaging

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/logger"
	"golang.org/x/net/proxy"

	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

const (
	cstDefaultTimeout time.Duration = 15 * time.Second
)

var (
	ErrInvalidSubject = errors.New("Invalid Subject")
	ErrInvalidHandler = errors.New("Invalid Handler")

	ErrRequest  = errors.New("Request error")
	ErrEncoding = errors.New("Encoding error")
	ErrDecoding = errors.New("Decoding error")
)

type Nats struct {
	nc *nats.Conn
}

type customDialer struct {
	ctx             context.Context
	connectTimeout  time.Duration
	connectTimeWait time.Duration
}

func (cd *customDialer) Dial(network, address string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(cd.ctx, cd.connectTimeout)
	defer cancel()

	for {
		logrus.Println("Attempting to connect to", address)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		select {
		case <-cd.ctx.Done():
			return nil, cd.ctx.Err()
		default:
			dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9150", nil, nil)
			if err != nil {
				panic(err)
			}

			if conn, err := dialer.Dial(network, address); err == nil {
				logrus.Println("Connected to NATS successfully")
				return conn, nil
			} else {
				time.Sleep(cd.connectTimeWait)
			}
		}
	}
}

// NewNats returns Nats messaging.
// panic on connection error
func NewNats(ctx context.Context, options NatsOptions) *Nats {
	log := logger.Logger(ctx).WithField("Method", "messaging.NewNats")
	url := fmt.Sprintf("nats://%s:%d", options.HostName, options.Port)

	withTor := strings.HasSuffix(options.HostName, ".onion")

	var err error
	var nc *nats.Conn

	var cd *customDialer
	if withTor {
		cd = &customDialer{
			ctx:             ctx,
			connectTimeout:  10 * time.Second,
			connectTimeWait: 1 * time.Second,
		}
	}
	opts := []nats.Option{
		nats.ReconnectWait(2 * time.Second),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Println("Reconnected to", c.ConnectedUrl())
		}),
		nats.DisconnectHandler(func(c *nats.Conn) {
			log.Println("Disconnected from NATS")
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Println("NATS connection is closed.")
		}),
		// nats.NoReconnect(),
	}
	if cd != nil {
		opts = append([]nats.Option{nats.SetCustomDialer(cd)}, opts...)
	}
	nc, err = nats.Connect(url, opts...)
	if err != nil {
		log.WithError(err).
			WithField("URL", url).
			Panic("Nats Connect failed")
	}
	return &Nats{
		nc: nc,
	}
}

func (n *Nats) NC() bank.NC {
	return n.nc
}

func natsMessageHandler(ctx context.Context, log *logrus.Entry, msg *nats.Msg, handle bank.MessageHandler) {
	log.
		WithField("Subject", msg.Subject).
		WithField("DataLength", len(msg.Data)).
		Trace("Handling nats message")

		// retrieve request
	req := new(bank.Message)
	err := req.Decode(msg.Data)
	if err != nil {
		log.
			WithError(err).
			Error("Failed to decode request")
		return
	}

	resp, err := handle(ctx, msg.Subject, req)
	if err != nil {
		log.
			WithError(err).
			Error("Request handling failed")
		// continue if reply is needed
	}

	// check if reply if requested
	if len(msg.Reply) == 0 {
		return
	}

	// prepare response
	// response can be nil if handler return and error
	if resp == nil {
		resp = bank.NewMessage()
		resp.Error = fmt.Sprintf("%s", err)
	}
	data, err := resp.Encode()
	if err != nil {
		log.
			WithError(err).
			Error("Failed to encode response")
		return
	}

	// send response to requester
	err = msg.Respond(data)
	if err != nil {
		log.
			WithError(err).
			Error("Failed to send response")
		return
	}
}

func clamp(count, min, max int) int {
	if count < min {
		return min
	} else if count > max {
		return max
	} else {
		return count
	}
}

// SubscribeWorkers
func (n *Nats) SubscribeWorkers(ctx context.Context, subject string, workerCount int, handle bank.MessageHandler) {
	workerCount = clamp(workerCount, 1, 1024)
	for w := 0; w < workerCount; w++ {
		n.Subscribe(ctx, subject, handle)
	}
}

// Subscribe
func (n *Nats) Subscribe(ctx context.Context, subject string, handle bank.MessageHandler) {
	log := logger.Logger(ctx).WithField("Method", "messaging.Nats.Subscribe")

	if len(subject) == 0 {
		log.
			WithError(ErrInvalidSubject).
			Panic("Invalid subject")
	}
	if handle == nil {
		log.WithError(ErrInvalidHandler).
			Panic("Invalid handler")
	}

	_, err := n.nc.QueueSubscribe(subject, subject+"_workers",
		func(msg *nats.Msg) {
			natsMessageHandler(ctx, log, msg, handle)
		},
	)
	if err != nil {
		log.
			WithError(err).
			WithField("subject", subject).
			Panic("Nats QueueSubscribe failed")
	}
}

// Publish perform nats Publish with subject and message.
// panic if subject or message are invalid
func (n *Nats) Publish(ctx context.Context, subject string, message *bank.Message) error {
	log := logger.Logger(ctx).WithField("Method", "messaging.Nats.Publish")

	if len(subject) == 0 {
		log.WithError(ErrInvalidSubject).
			Panic("Invalid subject")
	}
	if message == nil {
		log.WithError(bank.ErrInvalidMessage).
			Panic("Invalid message")
	}

	// prepare request
	data, err := message.Encode()
	if err != nil {
		log.WithError(err).
			Debug("Failed to encode message")
		return ErrEncoding
	}

	// perform nats publish
	err = n.nc.Publish(subject, data)
	if err != nil {
		log.WithError(err).
			Debug("Nats Publish failed")
		return ErrRequest
	}
	return nil
}

// Request perform nats Request with subject and message.
// use default timout
// panic if subject or message are invalid
func (n *Nats) Request(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	return n.RequestWithTimeout(ctx, subject, message, cstDefaultTimeout)
}

// RequestWithTimeout perform nats Request with subject and message.
// panic if subject or message are invalid
func (n *Nats) RequestWithTimeout(ctx context.Context, subject string, message *bank.Message, timeout time.Duration) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "messaging.Nats.RequestWithTimeout")

	if len(subject) == 0 {
		log.WithError(ErrInvalidSubject).
			Panic("Invalid subject")
	}
	if message == nil {
		log.WithError(bank.ErrInvalidMessage).
			Panic("Invalid message")
	}

	// prepare request
	data, err := message.Encode()
	if err != nil {
		log.WithError(err).
			Debug("Failed to encode message")
		return nil, ErrEncoding
	}

	// perform nats request
	msg, err := n.nc.Request(subject, data, timeout)
	if err != nil {
		log.WithError(err).
			Debug("Nats Request failed")
		return nil, ErrRequest
	}

	// retrieve response
	resp := new(bank.Message)
	err = resp.Decode(msg.Data)
	if err != nil {
		log.WithError(err).
			Debug("Failed to decode response")
		return nil, ErrDecoding
	}
	return resp, nil
}
