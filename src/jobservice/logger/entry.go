package logger

import "fmt"

// Entry provides unique interfaces on top of multiple logger backends.
// Entry also implements @Interface.
type Entry struct {
	loggers []Interface
}

// NewEntry creates a new logger Entry
func NewEntry(loggers []Interface) *Entry {
	return &Entry{
		loggers: loggers,
	}
}

// Debug ...
func (e *Entry) Debug(v ...interface{}) {
	for _, l := range e.loggers {
		l.Debug(v...)
	}
}

// Debugf with format
func (e *Entry) Debugf(format string, v ...interface{}) {
	for _, l := range e.loggers {
		l.Debugf(format, v...)
	}
}

// Info ...
func (e *Entry) Info(v ...interface{}) {
	for _, l := range e.loggers {
		l.Info(v...)
	}
}

// Infof with format
func (e *Entry) Infof(format string, v ...interface{}) {
	for _, l := range e.loggers {
		l.Infof(format, v...)
	}
}

// Warning ...
func (e *Entry) Warning(v ...interface{}) {
	for _, l := range e.loggers {
		l.Warning(v...)
	}
}

// Warningf with format
func (e *Entry) Warningf(format string, v ...interface{}) {
	for _, l := range e.loggers {
		l.Warningf(format, v...)
	}
}

// Error ...
func (e *Entry) Error(v ...interface{}) {
	for _, l := range e.loggers {
		l.Error(v...)
	}
}

// Errorf with format
func (e *Entry) Errorf(format string, v ...interface{}) {
	for _, l := range e.loggers {
		l.Errorf(format, v...)
	}
}

// Fatal error
func (e *Entry) Fatal(v ...interface{}) {
	for _, l := range e.loggers {
		l.Fatal(v...)
	}
}

// Fatalf error
func (e *Entry) Fatalf(format string, v ...interface{}) {
	for _, l := range e.loggers {
		l.Fatalf(format, v...)
	}
}

// Close logger
func (e *Entry) Close() error {
	var errMsg string
	for _, l := range e.loggers {
		if closer, ok := l.(Closer); ok {
			err := closer.Close()
			if err != nil {
				if errMsg == "" {
					errMsg = fmt.Sprintf("logger: %s, err: %s", GetLoggerName(l), err)
				} else {
					errMsg = fmt.Sprintf("%s; logger: %s, err: %s", errMsg, GetLoggerName(l), err)
				}
			}
		}
	}
	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}
	return nil
}
