package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/utils"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

const (
	Version = "0.1"
)

type MonitorApi int

func (p *MonitorApi) Run(ctx context.Context, port int) {
	log := logger.Logger(ctx).WithField("Method", "monitor.MonitorApi.Run")

	muxer := http.NewServeMux()
	muxer.HandleFunc("/getreport", Report)

	handler := negroni.New(&negroni.Recovery{})
	handler.UseHandler(muxer)

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
