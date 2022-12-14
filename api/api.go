package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"git.condensat.tech/bank/api/common"
	"git.condensat.tech/bank/api/handlers"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"git.condensat.tech/bank/api/oauth"
	"git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/api/sessions"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type Api int

func (p *Api) Run(ctx context.Context, port int, corsAllowedOrigins []string, oauthOptions oauth.Options) {
	log := logger.Logger(ctx).WithField("Method", "api.Api.Run")

	// create session and and to context
	session := sessions.NewSession(ctx)
	ctx = context.WithValue(ctx, sessions.KeySessions, session)
	// Add Domain to context
	if len(oauthOptions.Domain) > 0 {
		ctx = appcontext.WithDomain(ctx, oauthOptions.Domain)
	}

	p.registerHandlers(ctx)

	err := oauth.Init(oauthOptions)
	if err != nil {
		log.WithError(err).
			Warning("OAuth Init failed")
	}
	muxer := mux.NewRouter()

	services.RegisterMessageHandlers(ctx)
	services.RegisterServices(ctx, muxer, corsAllowedOrigins)

	oauth.RegisterHandlers(ctx, muxer)

	handler := negroni.New(&negroni.Recovery{})
	handler.Use(services.StatsMiddleware)
	handler.UseFunc(MiddlewarePeerRateLimiter)
	handler.UseFunc(AddWorkerHeader)
	handler.UseFunc(AddWorkerVersion)
	handler.UseHandler(muxer)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 16, // 16 KiB
		ConnContext:    func(conCtx context.Context, c net.Conn) context.Context { return ctx },
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.WithError(err).
				Info("Http server exited")
		}
	}()

	log.WithFields(logrus.Fields{
		"Hostname": utils.Hostname(),
		"Port":     port,
	}).Info("Api Service started")

	<-ctx.Done()
}

func (p *Api) registerHandlers(ctx context.Context) {
	log := logger.Logger(ctx).WithField("Method", "APi.RegisterHandlers")

	nats := appcontext.Messaging(ctx)

	const concurencyLevel = 2

	nats.SubscribeWorkers(ctx, common.UserCreateSubject, concurencyLevel, handlers.OnUserCreate)

	log.Debug("Bank Api registered")
}

// AddWorkerHeader - adds header of which node actually processed request
func AddWorkerHeader(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker", utils.Hostname())
	next(rw, r)
}

// AddWorkerVersion - adds header of which version is installed
func AddWorkerVersion(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker-Version", services.Version)
	next(rw, r)
}
