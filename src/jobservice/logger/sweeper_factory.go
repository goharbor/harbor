package logger

import (
	"errors"

	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
)

// SweeperFactory is responsible for creating a sweeper.Interface based on the settings
type SweeperFactory func(options ...OptionItem) (sweeper.Interface, error)

// FileSweeperFactory creates file sweeper.
func FileSweeperFactory(options ...OptionItem) (sweeper.Interface, error) {
	var workDir, duration = "", 1
	for _, op := range options {
		switch op.Field() {
		case "work_dir":
			workDir = op.String()
		case "duration":
			if op.Int() > 0 {
				duration = op.Int()
			}
		default:
		}
	}

	if len(workDir) == 0 {
		return nil, errors.New("missing required option 'work_dir'")
	}

	return sweeper.NewFileSweeper(workDir, duration), nil
}

// DBSweeperFactory creates DB sweeper.
func DBSweeperFactory(options ...OptionItem) (sweeper.Interface, error) {
	var duration = 1
	for _, op := range options {
		switch op.Field() {
		case "duration":
			if op.Int() > 0 {
				duration = op.Int()
			}
		default:
		}
	}

	return sweeper.NewDBSweeper(duration), nil
}
