package services

import (
	"context"
	"errors"
	"net/http"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/api/sessions"
	"git.condensat.tech/bank/appcontext"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
)

var (
	ErrServiceInternalError = errors.New("Service Internal Error")
)

func RegisterServices(ctx context.Context, mux *http.ServeMux, corsAllowedOrigins []string) {
	corsHandler := CreateCorsOptions(corsAllowedOrigins)

	mux.Handle("/api/v1/session", corsHandler.Handler(NewSessionHandler()))
}

func NewSessionHandler() http.Handler {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterService(new(SessionService), "session")

	return server
}

func ContextValues(ctx context.Context) (db bank.Database, session *sessions.Session, err error) {
	db = appcontext.Database(ctx)
	if ctxSession, ok := ctx.Value(sessions.KeySessions).(*sessions.Session); ok {
		session = ctxSession
	}
	if db == nil || session == nil {
		db = nil
		session = nil
		err = ErrServiceInternalError
		return
	}

	return
}
