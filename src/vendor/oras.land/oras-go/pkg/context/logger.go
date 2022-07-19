/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
