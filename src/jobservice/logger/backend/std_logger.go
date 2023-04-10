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

package backend

import (
	"os"

	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// StdOut represents os.Stdout
	StdOut = "std_out"
	// StdErr represents os.StdErr
	StdErr = "std_err"
)

// StdOutputLogger is an implementation of logger.Interface.
// It outputs the log to the stdout/stderr.
type StdOutputLogger struct {
	backendLogger *log.Logger
}

// NewStdOutputLogger creates a new std output logger
func NewStdOutputLogger(level string, output string, depth int) *StdOutputLogger {
	logLevel := parseLevel(level)
	logStream := os.Stdout
	if output == StdErr {
		logStream = os.Stderr
	}
	backendLogger := log.New(logStream, log.NewTextFormatter(), logLevel, depth)

	return &StdOutputLogger{
		backendLogger: backendLogger,
	}
}

// Debug ...
func (sl *StdOutputLogger) Debug(v ...interface{}) {
	sl.backendLogger.Debug(v...)
}

// Debugf with format
func (sl *StdOutputLogger) Debugf(format string, v ...interface{}) {
	sl.backendLogger.Debugf(format, v...)
}

// Info ...
func (sl *StdOutputLogger) Info(v ...interface{}) {
	sl.backendLogger.Info(v...)
}

// Infof with format
func (sl *StdOutputLogger) Infof(format string, v ...interface{}) {
	sl.backendLogger.Infof(format, v...)
}

// Warning ...
func (sl *StdOutputLogger) Warning(v ...interface{}) {
	sl.backendLogger.Warning(v...)
}

// Warningf with format
func (sl *StdOutputLogger) Warningf(format string, v ...interface{}) {
	sl.backendLogger.Warningf(format, v...)
}

// Error ...
func (sl *StdOutputLogger) Error(v ...interface{}) {
	sl.backendLogger.Error(v...)
}

// Errorf with format
func (sl *StdOutputLogger) Errorf(format string, v ...interface{}) {
	sl.backendLogger.Errorf(format, v...)
}

// Fatal error
func (sl *StdOutputLogger) Fatal(v ...interface{}) {
	sl.backendLogger.Fatal(v...)
}

// Fatalf error
func (sl *StdOutputLogger) Fatalf(format string, v ...interface{}) {
	sl.backendLogger.Fatalf(format, v...)
}
