// Copyright Project Harbor Authors. All rights reserved.

package logger

// Interface for logger.
type Interface interface {
	// For debuging
	Debug(v ...interface{})

	// For debuging with format
	Debugf(format string, v ...interface{})

	// For logging info
	Info(v ...interface{})

	// For logging info with format
	Infof(format string, v ...interface{})

	// For warning
	Warning(v ...interface{})

	// For warning with format
	Warningf(format string, v ...interface{})

	// For logging error
	Error(v ...interface{})

	// For logging error with format
	Errorf(format string, v ...interface{})

	// For fatal error
	Fatal(v ...interface{})

	// For fatal error with error
	Fatalf(format string, v ...interface{})
}
