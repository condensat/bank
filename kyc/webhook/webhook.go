package webhook

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"git.condensat.tech/bank/api"
	"git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

const (
	Version          = "0.1"
	KeySynapsSecrets = "Api.Sessions"
)

type WebHook int

func (p *WebHook) Run(ctx context.Context, port int, secrets Secrets) {
	log := logger.Logger(ctx).WithField("Method", "webhook.WebHook.Run")

	muxer := http.NewServeMux()

	handler := negroni.New(&negroni.Recovery{})
	handler.Use(services.StatsMiddleware)
	handler.UseFunc(api.MiddlewarePeerRateLimiter)
	handler.UseFunc(AddWorkerHeader)
	handler.UseFunc(AddWorkerVersion)
	handler.UseHandler(muxer)

	ctx = context.WithValue(ctx, KeySynapsSecrets, &secrets)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
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

// AddWorkerHeader - adds header of which node actually processed request
func AddWorkerHeader(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker", utils.Hostname())
	next(rw, r)
}

// AddWorkerVersion - adds header of which version is installed
func AddWorkerVersion(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Add("X-Worker-Version", Version)
	next(rw, r)
}
