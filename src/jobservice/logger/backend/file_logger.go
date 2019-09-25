package backend

import (
	"os"

	"github.com/goharbor/harbor/src/common/utils/log"
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
