package services

import (
	"context"
	"errors"
	"net/http"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/api/sessions"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
)

var (
	ErrServiceInternalError = errors.New("Service Internal Error")
)

func RegisterMessageHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "RegisterMessageHandlers")

	nats := appcontext.Messaging(ctx)
	nats.SubscribeWorkers(ctx, VerifySessionSubject, 4, sessions.VerifySession)

	log.Debug("MessageHandlers registered")
}

func RegisterServices(ctx context.Context, mux *mux.Router, corsAllowedOrigins []string) {
	corsHandler := CreateCorsOptions(corsAllowedOrigins)

	mux.Handle("/api/v1/session", corsHandler.Handler(NewSessionHandler(ctx)))
	mux.Handle("/api/v1/user", corsHandler.Handler(NewUserHandler(ctx)))
	mux.Handle("/api/v1/accounting", corsHandler.Handler(NewAccountingHandler(ctx)))
	mux.Handle("/api/v1/wallet", corsHandler.Handler(NewWalletHandler(ctx)))
	mux.Handle("/api/v1/swap", corsHandler.Handler(NewSwapHandler(ctx)))
	mux.Handle("/api/v1/fiat", corsHandler.Handler(NewFiatHandler(ctx)))
}

func NewSessionHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(SessionService), "session")
	if err != nil {
		panic(err)
	}

	return server
}

func NewUserHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(UserService), "user")
	if err != nil {
		panic(err)
	}

	return server
}

func NewAccountingHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(AccountingService), "accounting")
	if err != nil {
		panic(err)
	}

	return server
}

func NewWalletHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(WalletService), "wallet")
	if err != nil {
		panic(err)
	}

	return server
}

func NewSwapHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(SwapService), "swap")
	if err != nil {
		panic(err)
	}

	return server
}

func NewFiatHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(FiatService), "fiat")
	if err != nil {
		panic(err)
	}

	return server
}

func ContextValues(ctx context.Context) (bank.Database, *sessions.Session, error) {
	db := appcontext.Database(ctx)
	session, err := sessions.ContextSession(ctx)
	if db == nil || session == nil {
		err = ErrServiceInternalError
	}

	if err != nil {
		return nil, nil, err
	}

	return db, session, nil
}
