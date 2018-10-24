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
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/logger"
	logging "github.com/op/go-logging"
)

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)
var moduleName = "JobService"

// JobLogger is an implementation of logger.Interface.
// It used in the job to output logs to the logfile.
type JobLogger struct {
	backendLogger *logging.Logger
	streamRef     *os.File
}

// New logger
func New(jobID string) (logger.Interface, error) {
	if len(jobID) == 0 {
		return nil, errors.New("no job ID is provided to initialize logger")
	}

	jobLogger := &JobLogger{}
	// Read logger settings from default config
	loggerSettings := config.DefaultConfig.LoggerConfig

	backends := []logging.Backend{}
	for _, logger := range loggerSettings {
		loggerPrefix := fmt.Sprintf("%s:%s", strings.ToUpper(logger.Kind), jobID)

		if logger.Kind == config.LoggerKindFile {
			logPath := path.Join(logger.BasePath, fmt.Sprintf("%s.log", jobID))
			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			jobLogger.streamRef = f

			fileBackend := logging.NewLogBackend(f, loggerPrefix, 0)
			fileFormatter := logging.NewBackendFormatter(fileBackend, format)
			fileLeveledBackend := logging.AddModuleLevel(fileFormatter)
			fileLeveledBackend.SetLevel(parseLevel(logger.LogLevel), moduleName)

			backends = append(backends, fileLeveledBackend)
			continue
		}

		// Should be STD outs
		stdOut := os.Stdout
		if logger.Kind == config.LoggerKindStdError {
			stdOut = os.Stderr
		}

		stdBackend := logging.NewLogBackend(stdOut, loggerPrefix, 0)
		stdFormatter := logging.NewBackendFormatter(stdBackend, format)
		stdLeveledBackend := logging.AddModuleLevel(stdFormatter)
		stdLeveledBackend.SetLevel(parseLevel(logger.LogLevel), moduleName)
		backends = append(backends, stdLeveledBackend)
	}

	logging.SetBackend(backends...)

	jobLogger.backendLogger = logging.MustGetLogger(moduleName)

	return jobLogger, nil
}

// Close the opened io stream
// Implements logger.Closer interface
func (jl *JobLogger) Close() error {
	if jl.streamRef != nil {
		return jl.streamRef.Close()
	}

	return nil
}

// Debug ...
func (jl *JobLogger) Debug(v ...interface{}) {
	jl.backendLogger.Debug(createValueFormat(len(v)), v...)
}

// Debugf with format
func (jl *JobLogger) Debugf(format string, v ...interface{}) {
	jl.backendLogger.Debugf(format, v...)
}

// Info ...
func (jl *JobLogger) Info(v ...interface{}) {
	jl.backendLogger.Info(createValueFormat(len(v)), v...)
}

// Infof with format
func (jl *JobLogger) Infof(format string, v ...interface{}) {
	jl.backendLogger.Infof(format, v...)
}

// Warning ...
func (jl *JobLogger) Warning(v ...interface{}) {
	jl.backendLogger.Warning(createValueFormat(len(v)), v...)
}

// Warningf with format
func (jl *JobLogger) Warningf(format string, v ...interface{}) {
	jl.backendLogger.Warningf(format, v...)
}

// Error ...
func (jl *JobLogger) Error(v ...interface{}) {
	jl.backendLogger.Error(createValueFormat(len(v)), v...)
}

// Errorf with format
func (jl *JobLogger) Errorf(format string, v ...interface{}) {
	jl.backendLogger.Errorf(format, v...)
}

// Fatal error
func (jl *JobLogger) Fatal(v ...interface{}) {
	jl.backendLogger.Critical(createValueFormat(len(v)), v...)
}

// Fatalf error
func (jl *JobLogger) Fatalf(format string, v ...interface{}) {
	jl.backendLogger.Critical(format, v...)
}

func parseLevel(lvl string) logging.Level {

	var level = logging.INFO

	switch strings.ToLower(lvl) {
	case "debug":
		level = logging.DEBUG
	case "info":
		level = logging.INFO
	case "warning":
		level = logging.WARNING
	case "error":
		level = logging.ERROR
	case "fatal":
		level = logging.CRITICAL
	default:
	}

	return level
}

func createValueFormat(count int) string {
	f := []string{}
	for i := 0; i < count; i++ {
		f = append(f, "%s")
	}

	if len(f) == 0 {
		return ""
	}

	return strings.Join(f, ";")
}
