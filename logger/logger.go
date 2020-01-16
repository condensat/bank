package logger

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

const (
	appKey = iota
	databaseKey
	writerKey
	logLevelKey
)

// WithAppName returns a context with the Application name set
func WithAppName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, appKey, name)
}

// WithWriter returns a context with the log Writer set
func WithDatabase(ctx context.Context, database *DatabaseLogger) context.Context {
	return context.WithValue(ctx, databaseKey, database)
}

// WithWriter returns a context with the log Writer set
func WithWriter(ctx context.Context, writer io.Writer) context.Context {
	return context.WithValue(ctx, writerKey, writer)
}

// WithLogLevel returns a context with the LogLevel set
func WithLogLevel(ctx context.Context, level string) context.Context {
	return context.WithValue(ctx, logLevelKey, level)
}

func Logger(ctx context.Context) *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(contextLevel(ctx))
	logger.SetFormatter(&logrus.JSONFormatter{})
	entry := logrus.NewEntry(logger)

	if ctx == nil {
		return entry
	}

	if ctxApp, ok := ctx.Value(appKey).(string); ok {
		entry = entry.WithField("app", ctxApp)
	}
	if ctxWriter, ok := ctx.Value(writerKey).(io.Writer); ok {
		logger.SetOutput(ctxWriter)
	}
	if ctxLogLevel, ok := ctx.Value(logLevelKey).(string); ok {
		if level, err := logrus.ParseLevel(ctxLogLevel); err == nil {
			logger.SetLevel(level)
		}
	}

	return entry
}

func contextLevel(ctx context.Context) logrus.Level {
	if ctxLogLevel, ok := ctx.Value(logLevelKey).(string); ok {
		if level, err := logrus.ParseLevel(ctxLogLevel); err == nil {
			return level
		}
	}
	return logrus.WarnLevel
}

func contextDatabase(ctx context.Context) *DatabaseLogger {
	if ctxDatabase, ok := ctx.Value(databaseKey).(*DatabaseLogger); ok {
		return ctxDatabase
	}
	return nil
}
