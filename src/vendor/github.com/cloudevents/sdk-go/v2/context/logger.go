/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"context"

	"go.uber.org/zap"
)

// Opaque key type used to store logger
type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

// fallbackLogger is the logger is used when there is no logger attached to the context.
var fallbackLogger *zap.SugaredLogger

func init() {
	if logger, err := zap.NewProduction(); err != nil {
		// We failed to create a fallback logger.
		fallbackLogger = zap.NewNop().Sugar()
	} else {
		fallbackLogger = logger.Named("fallback").Sugar()
	}
}

// WithLogger returns a new context with the logger injected into the given context.
func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	if logger == nil {
		return context.WithValue(ctx, loggerKey, fallbackLogger)
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFrom returns the logger stored in context.
func LoggerFrom(ctx context.Context) *zap.SugaredLogger {
	l := ctx.Value(loggerKey)
	if l != nil {
		if logger, ok := l.(*zap.SugaredLogger); ok {
			return logger
		}
	}
	return fallbackLogger
}
