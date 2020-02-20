package monitor

import (
	"net/http"

	"git.condensat.tech/bank/logger"
)

func Report(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := logger.Logger(ctx).WithField("Method", "monitor.MonitorApi.Report")

	_, err := w.Write([]byte(jsonReportMock))
	if err != nil {
		log.WithError(err).
			Error("ResponseWriter Write failed")

	}

	log.Debug("Report sent")
}
