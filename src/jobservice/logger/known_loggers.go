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

package logger

import (
	"reflect"
	"slices"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/logger/backend"
)

const (
	// NameFile is unique name of the file logger.
	NameFile = "FILE"
	// NameStdOutput is the unique name of the std logger.
	NameStdOutput = "STD_OUTPUT"
	// NameDB is the unique name of the DB logger.
	NameDB = "DB"
)

// Declaration is used to declare a supported logger.
// Use this declaration to indicate what logger and sweeper will be provided.
type Declaration struct {
	Logger  Factory
	Sweeper SweeperFactory
	Getter  GetterFactory
	// Indicate if the logger is a singleton logger
	Singleton bool
}

// knownLoggers is a static logger registry.
// All the implemented loggers (w/ sweeper) should be registered
// with an unique name in this registry. Then they can be used to
// log info.
var knownLoggers = map[string]*Declaration{
	// File logger
	NameFile: {FileFactory, FileSweeperFactory, FileGetterFactory, false},
	// STD output(both stdout and stderr) logger
	NameStdOutput: {StdFactory, nil, nil, true},
	// DB logger
	NameDB: {DBFactory, DBSweeperFactory, DBGetterFactory, false},
}

// IsKnownLogger checks if the logger is supported with name.
func IsKnownLogger(name string) (*Declaration, bool) {
	d, ok := knownLoggers[name]

	return d, ok
}

// HasSweeper checks if the logger with the name provides a sweeper.
func HasSweeper(name string) bool {
	d, ok := knownLoggers[name]

	return ok && d.Sweeper != nil
}

// HasGetter checks if the logger with the name provides a log data getter.
func HasGetter(name string) bool {
	d, ok := knownLoggers[name]

	return ok && d.Getter != nil
}

// All known levels which are supported.
var debugLevels = []string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"FATAL",
}

// IsKnownLevel is used to check if the logger level is supported.
func IsKnownLevel(level string) bool {
	if len(level) == 0 {
		return false
	}

	return slices.Contains(debugLevels, strings.ToUpper(level))
}

// GetLoggerName return a logger name by Interface
func GetLoggerName(l Interface) string {
	var name string
	if l == nil {
		return name
	}

	switch l.(type) {
	case *backend.DBLogger:
		name = NameDB
	case *backend.StdOutputLogger:
		name = NameStdOutput
	case *backend.FileLogger:
		name = NameFile
	default:
		name = reflect.TypeOf(l).String()
	}

	return name
}
