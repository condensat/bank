package processus

import (
	"context"
	"os"
	"runtime"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"
	"git.condensat.tech/bank/monitor"
	"git.condensat.tech/bank/monitor/common"
	"git.condensat.tech/bank/monitor/messaging"
	"git.condensat.tech/bank/utils"
)

type Grabber struct {
	appName   string
	interval  time.Duration
	messaging bank.Messaging
}

func NewGrabber(ctx context.Context, interval time.Duration) *Grabber {
	return &Grabber{
		appName:   appcontext.AppName(ctx),
		interval:  interval,
		messaging: appcontext.Messaging(ctx),
	}
}

func (p *Grabber) Run(ctx context.Context, numWorkers int) {
	log := logger.Logger(ctx).WithField("Method", "processus.Grabber.Run")

	var clock monitor.Clock
	for {
		clock.Init()
		select {
		case <-time.After(p.interval):
			processInfo := processInfo(p.appName, &clock)
			err := p.sendProcessInfo(ctx, &processInfo)
			if err != nil {
				log.WithError(err).Error("Failed to sendProcessInfo")
				continue
			}
			log.Trace("Grab processInfo")

		case <-ctx.Done():
			log.Info("Process Grabber done.")
			return
		}
	}
}

func processInfo(appName string, clock *monitor.Clock) common.ProcessInfo {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return common.ProcessInfo{
		Timestamp: time.Now().UTC().Truncate(time.Second),
		AppName:   appName,
		Hostname:  utils.Hostname(),
		PID:       os.Getpid(),

		MemAlloc:      mem.Alloc,
		MemTotalAlloc: mem.TotalAlloc,
		MemSys:        mem.Sys,
		MemLookups:    mem.Lookups,

		NumCPU:       uint64(runtime.NumCPU()),
		NumGoroutine: uint64(runtime.NumGoroutine()),
		NumCgoCall:   uint64(runtime.NumCgoCall()),
		CPUUsage:     clock.CPU(),
	}
}

func (p *Grabber) sendProcessInfo(ctx context.Context, processInfo *common.ProcessInfo) error {
	request := bank.ToMessage(p.appName, processInfo)
	return p.messaging.Publish(ctx, messaging.InboundSubject, request)
}
