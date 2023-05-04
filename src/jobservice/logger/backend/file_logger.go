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

// FileLogger is an implementation of logger.Interface.
// It outputs logs to the specified logfile.
type FileLogger struct {
	backendLogger *log.Logger
	streamRef     *os.File
}

// NewFileLogger crates a new file logger
// nil might be returned
func NewFileLogger(level string, logPath string, depth int) (*FileLogger, error) {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	logLevel := parseLevel(level)
	backendLogger := log.New(f, log.NewTextFormatter(), logLevel, depth)

	return &FileLogger{
		backendLogger: backendLogger,
		streamRef:     f,
	}, nil
}

// Close the opened io stream
// Implements logger.Closer interface
func (fl *FileLogger) Close() error {
	if fl.streamRef != nil {
		return fl.streamRef.Close()
	}

	return nil
}

// Debug ...
func (fl *FileLogger) Debug(v ...interface{}) {
	fl.backendLogger.Debug(v...)
}

// Debugf with format
func (fl *FileLogger) Debugf(format string, v ...interface{}) {
	fl.backendLogger.Debugf(format, v...)
}

// Info ...
func (fl *FileLogger) Info(v ...interface{}) {
	fl.backendLogger.Info(v...)
}

// Infof with format
func (fl *FileLogger) Infof(format string, v ...interface{}) {
	fl.backendLogger.Infof(format, v...)
}

// Warning ...
func (fl *FileLogger) Warning(v ...interface{}) {
	fl.backendLogger.Warning(v...)
}

// Warningf with format
func (fl *FileLogger) Warningf(format string, v ...interface{}) {
	fl.backendLogger.Warningf(format, v...)
}

// Error ...
func (fl *FileLogger) Error(v ...interface{}) {
	fl.backendLogger.Error(v...)
}

// Errorf with format
func (fl *FileLogger) Errorf(format string, v ...interface{}) {
	fl.backendLogger.Errorf(format, v...)
}

// Fatal error
func (fl *FileLogger) Fatal(v ...interface{}) {
	fl.backendLogger.Fatal(v...)
}

// Fatalf error
func (fl *FileLogger) Fatalf(format string, v ...interface{}) {
	fl.backendLogger.Fatalf(format, v...)
}
