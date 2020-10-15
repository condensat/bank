package messaging

import (
	nats "github.com/nats-io/nats.go"
)

func ToNats(messaging Messaging) *nats.Conn {
	nc := messaging.NC()
	return nc.(*nats.Conn)
}
