package logger

import (
	"os"
	"strings"

	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice/logger"
)

//JobLogger is an implementation of logger.Interface.
//It used in the job to output logs to the logfile.
type JobLogger struct {
	backendLogger *log.Logger
	streamRef     *os.File
}

//New logger
//nil might be returned
func New(logPath string, level string) logger.Interface {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}
	logLevel := parseLevel(level)
	backendLogger := log.New(f, log.NewTextFormatter(), logLevel)

	return &JobLogger{
		backendLogger: backendLogger,
		streamRef:     f,
	}
}

//Close the opened io stream
//Implements logger.Closer interface
func (jl *JobLogger) Close() error {
	if jl.streamRef != nil {
		return jl.streamRef.Close()
	}

	return nil
}

//Debug ...
func (jl *JobLogger) Debug(v ...interface{}) {
	jl.backendLogger.Debug(v...)
}

//Debugf with format
func (jl *JobLogger) Debugf(format string, v ...interface{}) {
	jl.backendLogger.Debugf(format, v...)
}

//Info ...
func (jl *JobLogger) Info(v ...interface{}) {
	jl.backendLogger.Info(v...)
}

//Infof with format
func (jl *JobLogger) Infof(format string, v ...interface{}) {
	jl.backendLogger.Infof(format, v...)
}

//Warning ...
func (jl *JobLogger) Warning(v ...interface{}) {
	jl.backendLogger.Warning(v...)
}

//Warningf with format
func (jl *JobLogger) Warningf(format string, v ...interface{}) {
	jl.backendLogger.Warningf(format, v...)
}

//Error ...
func (jl *JobLogger) Error(v ...interface{}) {
	jl.backendLogger.Error(v...)
}

//Errorf with format
func (jl *JobLogger) Errorf(format string, v ...interface{}) {
	jl.backendLogger.Errorf(format, v...)
}

//Fatal error
func (jl *JobLogger) Fatal(v ...interface{}) {
	jl.backendLogger.Fatal(v...)
}

//Fatalf error
func (jl *JobLogger) Fatalf(format string, v ...interface{}) {
	jl.backendLogger.Fatalf(format, v...)
}

func parseLevel(lvl string) log.Level {

	var level = log.WarningLevel

	switch strings.ToLower(lvl) {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warning":
		level = log.WarningLevel
	case "error":
		level = log.ErrorLevel
	case "fatal":
		level = log.FatalLevel
	default:
	}

	return level
}
