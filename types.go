package bank

import (
	"context"

	"git.condensat.tech/bank/security/secureid"
)

type ServerOptions struct {
	Protocol string
	HostName string
	Port     int
}

type Worker interface {
	Run(ctx context.Context, numWorkers int)
}

type SecureID interface {
	ToSecureID(context string, value secureid.Value) (secureid.SecureID, error)
	FromSecureID(context string, secureID secureid.SecureID) (secureid.Value, error)

	ToString(secureID secureid.SecureID) string
	Parse(secureID string) secureid.SecureID
}
