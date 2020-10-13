package sessions

import (
	"context"
	"errors"
)

var (
	ErrInternalError = errors.New("Internal Error")
)

func ContextSession(ctx context.Context) (*Session, error) {
	if ctxSession, ok := ctx.Value(KeySessions).(*Session); ok {
		return ctxSession, nil
	}
	return nil, ErrInternalError
}
