package logger

import (
	"errors"

	"github.com/goharbor/harbor/src/jobservice/logger/getter"
)

// GetterFactory is responsible for creating a log data getter based on the options
type GetterFactory func(options ...OptionItem) (getter.Interface, error)

// FileGetterFactory creates a getter for the "FILE" logger
func FileGetterFactory(options ...OptionItem) (getter.Interface, error) {
	var baseDir string
	for _, op := range options {
		if op.Field() == "base_dir" {
			baseDir = op.String()
			break
		}
	}

	if len(baseDir) == 0 {
		return nil, errors.New("missing required option 'base_dir'")
	}

	return getter.NewFileGetter(baseDir), nil
}

// DBGetterFactory creates a getter for the DB logger
func DBGetterFactory(options ...OptionItem) (getter.Interface, error) {
	return getter.NewDBGetter(), nil
}
