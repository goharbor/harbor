// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"
)

var (
	// G shortcut to get logger from the context
	G = GetLogger
	// L the default logger
	L = DefaultLogger()
)

type loggerKey struct{}

// WithLogger returns a new context with the provided logger.
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger retrieves the current logger from the context.
// If no logger is available, the default logger is returned.
func GetLogger(ctx context.Context) *Logger {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return L
	}

	return logger.(*Logger)
}
