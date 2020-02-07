package messaging

import (
	"git.condensat.tech/bank"

	nats "github.com/nats-io/nats.go"
)

func ToNats(messaging bank.Messaging) *nats.Conn {
	nc := messaging.NC()
	return nc.(*nats.Conn)
}
