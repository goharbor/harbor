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
	"os"

	"github.com/op/go-logging"
)

// ServiceLogger is an implementation of logger.Interface.
// It used to log info in workerpool components.
type ServiceLogger struct {
	backendLogger *logging.Logger
}

// NewServiceLogger to create new logger for job service
// nil might be returned
func NewServiceLogger(level string) *ServiceLogger {
	stdBackend := logging.NewLogBackend(os.Stdout, "[JobService]", 0)
	stdFormatter := logging.NewBackendFormatter(stdBackend, format)
	stdLeveledBackend := logging.AddModuleLevel(stdFormatter)
	stdLeveledBackend.SetLevel(parseLevel(level), moduleName)
	logging.SetBackend(stdLeveledBackend)

	return &ServiceLogger{
		backendLogger: logging.MustGetLogger(moduleName),
	}
}

// Debug ...
func (sl *ServiceLogger) Debug(v ...interface{}) {
	sl.backendLogger.Debug(createValueFormat(len(v)), v...)
}

// Debugf with format
func (sl *ServiceLogger) Debugf(format string, v ...interface{}) {
	sl.backendLogger.Debugf(format, v...)
}

// Info ...
func (sl *ServiceLogger) Info(v ...interface{}) {
	sl.backendLogger.Info(createValueFormat(len(v)), v...)
}

// Infof with format
func (sl *ServiceLogger) Infof(format string, v ...interface{}) {
	sl.backendLogger.Infof(format, v...)
}

// Warning ...
func (sl *ServiceLogger) Warning(v ...interface{}) {
	sl.backendLogger.Warning(createValueFormat(len(v)), v...)
}

// Warningf with format
func (sl *ServiceLogger) Warningf(format string, v ...interface{}) {
	sl.backendLogger.Warningf(format, v...)
}

// Error ...
func (sl *ServiceLogger) Error(v ...interface{}) {
	sl.backendLogger.Error(createValueFormat(len(v)), v...)
}

// Errorf with format
func (sl *ServiceLogger) Errorf(format string, v ...interface{}) {
	sl.backendLogger.Errorf(format, v...)
}

// Fatal error
func (sl *ServiceLogger) Fatal(v ...interface{}) {
	sl.backendLogger.Critical(createValueFormat(len(v)), v...)
}

// Fatalf error
func (sl *ServiceLogger) Fatalf(format string, v ...interface{}) {
	sl.backendLogger.Critical(format, v...)
}
