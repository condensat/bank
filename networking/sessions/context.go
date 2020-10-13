package sessions

import (
	"context"
)

func ContextSession(ctx context.Context) (*Session, error) {
	if ctxSession, ok := ctx.Value(KeySessions).(*Session); ok {
		return ctxSession, nil
	}
	return nil, ErrInternalError
}
