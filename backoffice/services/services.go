package services

import (
	"context"
	"net/http"

	"git.condensat.tech/bank/api/services"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
)

func RegisterServices(ctx context.Context, mux *mux.Router, corsAllowedOrigins []string) {
	corsHandler := services.CreateCorsOptions(corsAllowedOrigins)

	mux.Handle("/api/v1/dashboard", corsHandler.Handler(NewDashboardHandler(ctx)))
}

func NewDashboardHandler(ctx context.Context) http.Handler {
	server := rpc.NewServer()

	jsonCodec := services.NewCookieCodec(ctx)
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterCodec(jsonCodec, "application/json; charset=UTF-8") // For firefox 11 and other browsers which append the charset=UTF-8

	err := server.RegisterService(new(DashboardService), "dashboard")
	if err != nil {
		panic(err)
	}

	return server
}
