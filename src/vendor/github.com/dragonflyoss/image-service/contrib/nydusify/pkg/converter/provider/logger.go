// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// LoggerFields shows key-value like info in log line
type LoggerFields = map[string]interface{}

// ProgressLogger displays the progress log of conversion
type ProgressLogger interface {
	Log(ctx context.Context, msg string, fields LoggerFields) func(error) error
}

type defaultLogger struct{}

func (logger *defaultLogger) Log(ctx context.Context, msg string, fields LoggerFields) func(err error) error {
	if fields == nil {
		fields = make(LoggerFields)
	}
	logrus.WithFields(fields).Info(msg)
	start := time.Now()
	return func(err error) error {
		duration := time.Since(start)
		fields["Time"] = duration.String()
		logrus.WithFields(fields).Info(msg)
		return err
	}
}

// DefaultLogger provides a basic logger outputted to stdout
func DefaultLogger() (ProgressLogger, error) {
	return &defaultLogger{}, nil
}
