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
	"bufio"
	"bytes"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/joblog"
	"github.com/goharbor/harbor/src/pkg/joblog/models"
)

// DBLogger is an implementation of logger.Interface.
// It outputs logs to PGSql.
type DBLogger struct {
	backendLogger *log.Logger
	bw            *bufio.Writer
	buffer        *bytes.Buffer
	key           string
}

// NewDBLogger crates a new DB logger
// nil might be returned
func NewDBLogger(key string, level string, depth int) (*DBLogger, error) {
	buffer := bytes.NewBuffer(make([]byte, 0))
	bw := bufio.NewWriter(buffer)
	logLevel := parseLevel(level)

	backendLogger := log.New(bw, log.NewTextFormatter(), logLevel, depth)

	return &DBLogger{
		backendLogger: backendLogger,
		bw:            bw,
		buffer:        buffer,
		key:           key,
	}, nil
}

// Close the opened io stream and flush data into DB
// Implements logger.Closer interface
func (dbl *DBLogger) Close() error {
	err := dbl.bw.Flush()
	if err != nil {
		return err
	}

	jobLog := models.JobLog{
		UUID:    dbl.key,
		Content: dbl.buffer.String(),
	}

	_, err = joblog.Mgr.Create(orm.Context(), &jobLog)
	if err != nil {
		return err
	}
	return nil
}

// Debug ...
func (dbl *DBLogger) Debug(v ...any) {
	dbl.backendLogger.Debug(v...)
}

// Debugf with format
func (dbl *DBLogger) Debugf(format string, v ...any) {
	dbl.backendLogger.Debugf(format, v...)
}

// Info ...
func (dbl *DBLogger) Info(v ...any) {
	dbl.backendLogger.Info(v...)
}

// Infof with format
func (dbl *DBLogger) Infof(format string, v ...any) {
	dbl.backendLogger.Infof(format, v...)
}

// Warning ...
func (dbl *DBLogger) Warning(v ...any) {
	dbl.backendLogger.Warning(v...)
}

// Warningf with format
func (dbl *DBLogger) Warningf(format string, v ...any) {
	dbl.backendLogger.Warningf(format, v...)
}

// Error ...
func (dbl *DBLogger) Error(v ...any) {
	dbl.backendLogger.Error(v...)
}

// Errorf with format
func (dbl *DBLogger) Errorf(format string, v ...any) {
	dbl.backendLogger.Errorf(format, v...)
}

// Fatal error
func (dbl *DBLogger) Fatal(v ...any) {
	dbl.backendLogger.Fatal(v...)
}

// Fatalf error
func (dbl *DBLogger) Fatalf(format string, v ...any) {
	dbl.backendLogger.Fatalf(format, v...)
}
