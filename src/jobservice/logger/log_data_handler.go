package logger

import (
	"errors"

	"github.com/goharbor/harbor/src/jobservice/logger/getter"
)

var logDataGetter getter.Interface

// Retrieve is wrapper func for getter.Retrieve
func Retrieve(logID string) ([]byte, error) {
	if logDataGetter == nil {
		return nil, errors.New("no log data getter is configured")
	}

	return logDataGetter.Retrieve(logID)
}

// HasLogGetterConfigured checks if a log data getter is there for using
func HasLogGetterConfigured() bool {
	return logDataGetter != nil
}
