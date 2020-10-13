package services

import (
	"context"
	"errors"
	"net/http"

	"git.condensat.tech/bank/networking"
	"git.condensat.tech/bank/networking/sessions"

	"github.com/gorilla/rpc/v2"
)

var (
	ErrServiceInternalError = errors.New("Service Internal Error")
)

func RegisterServices(ctx context.Context, mux *http.ServeMux, corsAllowedOrigins []string) {
	corsHandler := networking.CreateCorsOptions(corsAllowedOrigins)

	mux.Handle("/api/v1/stack", corsHandler.Handler(NewStackHandler(ctx)))
}

func NewStackHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := sessions.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(StackService), "stack")
	if err != nil {
		panic(err)
	}

	return server
}
