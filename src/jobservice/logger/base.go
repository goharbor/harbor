package logger

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/logger/getter"
	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
)

const (
	systemKeyServiceLogger = "system.jobServiceLogger"
	systemKeyLogDataGetter = "system.logDataGetter"
)

var singletons sync.Map

// GetLogger gets an unified logger entry for logging per the passed settings.
// The logger may built based on the multiple registered logger backends.
//
//  loggerOptions ...Option : logger options
//
// If failed, a nil logger and a non-nil error will be returned.
// Otherwise, a non nil logger is returned with nil error.
func GetLogger(loggerOptions ...Option) (Interface, error) {
	lOptions := &options{
		values: make(map[string][]OptionItem),
	}

	// Config
	for _, op := range loggerOptions {
		op.Apply(lOptions)
	}

	// No options specified, enable std as default
	if len(loggerOptions) == 0 {
		defaultOp := BackendOption(NameStdOutput, "", nil)
		defaultOp.Apply(lOptions)
	}

	// Create backends
	loggers := make([]Interface, 0)
	for name, ops := range lOptions.values {
		if !IsKnownLogger(name) {
			return nil, fmt.Errorf("no logger registered for name '%s'", name)
		}

		d := KnownLoggers(name)
		var (
			l  Interface
			ok bool
		)

		// Singleton
		if d.Singleton {
			var li interface{}
			li, ok = singletons.Load(name)
			if ok {
				l = li.(Interface)
			}
		}

		if !ok {
			var err error
			l, err = d.Logger(ops...)
			if err != nil {
				return nil, err
			}

			// Cache it
			singletons.Store(name, l)
		}

		// Append to the logger list as logger entry backends
		loggers = append(loggers, l)
	}

	return NewEntry(loggers), nil
}

// GetSweeper gets an unified sweeper controller for sweeping purpose.
//
// context context.Context  : system context used to control the sweeping loops
// sweeperOptions ...Option : sweeper options
//
// If failed, a nil sweeper and a non-nil error will be returned.
// Otherwise, a non nil sweeper is returned with nil error.
func GetSweeper(context context.Context, sweeperOptions ...Option) (sweeper.Interface, error) {
	// No default sweeper will provide
	// If no one is configured, directly return nil with error
	if len(sweeperOptions) == 0 {
		return nil, errors.New("no options provided for creating sweeper controller")
	}

	sOptions := &options{
		values: make(map[string][]OptionItem),
	}

	// Config
	for _, op := range sweeperOptions {
		op.Apply(sOptions)
	}

	sweepers := make([]sweeper.Interface, 0)
	for name, ops := range sOptions.values {
		if !HasSweeper(name) {
			return nil, fmt.Errorf("no sweeper provided for the logger %s", name)
		}

		d := KnownLoggers(name)
		s, err := d.Sweeper(ops...)
		if err != nil {
			return nil, err
		}

		sweepers = append(sweepers, s)
	}

	return NewSweeperController(context, sweepers), nil
}

// GetLogDataGetter return the 1st matched log data getter interface
//
//  loggerOptions ...Option : logger options
//
// If failed,
//   configured but initialize failed: a nil log data getter and a non-nil error will be returned.
//   no getter configured: a nil log data getter with a nil error are returned
// Otherwise, a non nil log data getter is returned with nil error.
func GetLogDataGetter(loggerOptions ...Option) (getter.Interface, error) {
	if len(loggerOptions) == 0 {
		// If no options, directly return nil interface with error
		return nil, errors.New("no options provided to create log data getter")
	}

	lOptions := &options{
		values: make(map[string][]OptionItem),
	}

	// Config
	for _, op := range loggerOptions {
		op.Apply(lOptions)
	}

	// Iterate with specified order
	keys := make([]string, 0)
	for k := range lOptions.values {
		keys = append(keys, k)
	}

	// Sort
	sort.Strings(keys)

	for _, k := range keys {
		if HasGetter(k) {
			// 1st match
			d := knownLoggers[k]
			theGetter, err := d.Getter(lOptions.values[k]...)
			if err != nil {
				return nil, err
			}

			return theGetter, nil
		}
	}

	// No one configured
	return nil, nil
}

// Init the loggers and sweepers
func Init(ctx context.Context) error {
	// For loggers
	options := make([]Option, 0)
	// For sweepers
	sOptions := make([]Option, 0)

	for _, lc := range config.DefaultConfig.LoggerConfigs {
		// Inject logger depth here for FILE and STD logger to avoid configuring it in the yaml
		// For logger of job service itself, the depth should be 6
		if lc.Name == NameFile || lc.Name == NameStdOutput {
			if lc.Settings == nil {
				lc.Settings = map[string]interface{}{}
			}
			lc.Settings["depth"] = 6
		}
		options = append(options, BackendOption(lc.Name, lc.Level, lc.Settings))
		if lc.Sweeper != nil {
			sOptions = append(sOptions, SweeperOption(lc.Name, lc.Sweeper.Duration, lc.Sweeper.Settings))
		}
	}

	// Get loggers for job service
	lg, err := GetLogger(options...)
	if err != nil {
		return err
	}
	// Avoid data race issue
	singletons.Store(systemKeyServiceLogger, lg)

	jOptions := make([]Option, 0)
	// Append configured sweepers in job loggers if existing
	for _, lc := range config.DefaultConfig.JobLoggerConfigs {
		jOptions = append(jOptions, BackendOption(lc.Name, lc.Level, lc.Settings))
		if lc.Sweeper != nil {
			sOptions = append(sOptions, SweeperOption(lc.Name, lc.Sweeper.Duration, lc.Sweeper.Settings))
		}
	}

	// Get log data getter with the same options of job loggers
	g, err := GetLogDataGetter(jOptions...)
	if err != nil {
		return err
	}
	if g != nil {
		// Avoid data race issue
		singletons.Store(systemKeyLogDataGetter, g)
	}

	// If sweepers configured
	if len(sOptions) > 0 {
		// Get the sweeper controller
		swp, err := GetSweeper(ctx, sOptions...)
		if err != nil {
			return fmt.Errorf("create logger sweeper error: %s", err)
		}
		// Start sweep loop
		_, err = swp.Sweep()
		if err != nil {
			return fmt.Errorf("start logger sweeper error: %s", err)
		}
	}

	return nil
}
