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
func (dbl *DBLogger) Debug(v ...interface{}) {
	dbl.backendLogger.Debug(v...)
}

// Debugf with format
func (dbl *DBLogger) Debugf(format string, v ...interface{}) {
	dbl.backendLogger.Debugf(format, v...)
}

// Info ...
func (dbl *DBLogger) Info(v ...interface{}) {
	dbl.backendLogger.Info(v...)
}

// Infof with format
func (dbl *DBLogger) Infof(format string, v ...interface{}) {
	dbl.backendLogger.Infof(format, v...)
}

// Warning ...
func (dbl *DBLogger) Warning(v ...interface{}) {
	dbl.backendLogger.Warning(v...)
}

// Warningf with format
func (dbl *DBLogger) Warningf(format string, v ...interface{}) {
	dbl.backendLogger.Warningf(format, v...)
}

// Error ...
func (dbl *DBLogger) Error(v ...interface{}) {
	dbl.backendLogger.Error(v...)
}

// Errorf with format
func (dbl *DBLogger) Errorf(format string, v ...interface{}) {
	dbl.backendLogger.Errorf(format, v...)
}

// Fatal error
func (dbl *DBLogger) Fatal(v ...interface{}) {
	dbl.backendLogger.Fatal(v...)
}

// Fatalf error
func (dbl *DBLogger) Fatalf(format string, v ...interface{}) {
	dbl.backendLogger.Fatalf(format, v...)
}
