package context

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/containerd/containerd/log"
	"github.com/sirupsen/logrus"
)

// WithLogger returns a new context with the provided logger.
// This method wraps github.com/containerd/containerd/log.WithLogger()
func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return log.WithLogger(ctx, logger)
}

// WithLoggerFromWriter returns a new context with the logger, writting to the provided logger.
func WithLoggerFromWriter(ctx context.Context, writer io.Writer) context.Context {
	logger := logrus.New()
	logger.Out = writer
	entry := logrus.NewEntry(logger)
	return WithLogger(ctx, entry)
}

// WithLoggerDiscarded returns a new context with the logger, writting to nothing.
func WithLoggerDiscarded(ctx context.Context) context.Context {
	return WithLoggerFromWriter(ctx, ioutil.Discard)
}

// GetLogger retrieves the current logger from the context.
// This method wraps github.com/containerd/containerd/log.GetLogger()
func GetLogger(ctx context.Context) *logrus.Entry {
	return log.GetLogger(ctx)
}
