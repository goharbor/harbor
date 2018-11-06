package logger

import (
	"errors"

	"github.com/goharbor/harbor/src/jobservice/logger/getter"
)

// Retrieve is wrapper func for getter.Retrieve
func Retrieve(logID string) ([]byte, error) {
	val, ok := singletons.Load(systemKeyLogDataGetter)
	if !ok {
		return nil, errors.New("no log data getter is configured")
	}

	return val.(getter.Interface).Retrieve(logID)
}

// HasLogGetterConfigured checks if a log data getter is there for using
func HasLogGetterConfigured() bool {
	_, ok := singletons.Load(systemKeyLogDataGetter)
	return ok
}
