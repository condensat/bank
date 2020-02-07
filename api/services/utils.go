package services

import (
	"context"
	"net/http"
	"time"

	"git.condensat.tech/bank/logger"

	"github.com/sirupsen/logrus"
)

func makeTimestampMillis(ts time.Time) int64 {
	return ts.UnixNano() / int64(time.Millisecond)
}

func fromTimestampMillis(timestamp int64) time.Time {
	return time.Unix(0, int64(timestamp)*int64(time.Millisecond)).UTC()
}

func RequesterIP(r *http.Request) string {
	// Header added by reverse proxy
	const xRealIP = "X-Real-Ip"
	const xForwardedFor = "X-Forwarded-For"

	// Priority order
	if ips, ok := r.Header[xRealIP]; ok && len(ips) > 0 {
		return ips[0]
	} else if ips, ok := r.Header[xForwardedFor]; ok && len(ips) > 0 {
		return ips[0]
	} else {
		return r.RemoteAddr // fallback with RemoteAddr
	}
}

func AppendRequestLog(log *logrus.Entry, r *http.Request) *logrus.Entry {
	return log.WithFields(logrus.Fields{
		"UserAgent": r.UserAgent(),
		"IP":        RequesterIP(r),
		"URI":       r.RequestURI,
	})
}

func GetRequestLog(ctx context.Context, r *http.Request) *logrus.Entry {
	return AppendRequestLog(logger.Logger(ctx), r)
}

func GetServiceRequestLog(log *logrus.Entry, r *http.Request, service, operation string) *logrus.Entry {
	log = AppendRequestLog(log, r)

	// Optionals
	if len(service) > 0 {
		log = log.WithField("Service", service)
	}
	if len(service) > 0 {
		log = log.WithField("Operation", operation)
	}

	return log
}
