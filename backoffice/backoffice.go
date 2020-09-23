package backoffice

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"git.condensat.tech/bank/backoffice/services"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

type BackOffice int

func (p *BackOffice) Run(ctx context.Context, port int, corsAllowedOrigins []string) {
	log := logger.Logger(ctx).WithField("Method", "backoffice.BackOffice.Run")

	muxer := mux.NewRouter()

	services.RegisterServices(ctx, muxer, corsAllowedOrigins)

	handler := negroni.New(&negroni.Recovery{})
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
	}).Info("BackOffice Service started")

	<-ctx.Done()
}
