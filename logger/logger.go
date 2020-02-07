package logger

import (
	"context"
	"io"

	"git.condensat.tech/bank/appcontext"

	"github.com/sirupsen/logrus"
)

func Logger(ctx context.Context) *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(appcontext.Level(ctx))
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logrus.NewEntry(logger)

	if ctx == nil {
		return entry
	}

	if appName, ok := ctx.Value(appcontext.AppKey).(string); ok {
		entry = entry.WithField("app", appName)
	}
	if writerKey, ok := ctx.Value(appcontext.WriterKey).(io.Writer); ok {
		logger.SetOutput(writerKey)
	}
	if level, ok := ctx.Value(appcontext.LogLevelKey).(logrus.Level); ok {
		logger.SetLevel(level)
	}

	return entry
}
