// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

var logger = New(os.Stdout, NewTextFormatter(), WarningLevel)

func init() {
	logger.callDepth = 4

	lvl := os.Getenv("LOG_LEVEL")
	if len(lvl) == 0 {
		logger.SetLevel(InfoLevel)
		return
	}

	level, err := parseLevel(lvl)
	if err != nil {
		logger.SetLevel(InfoLevel)
		return
	}

	logger.SetLevel(level)

}

// Logger provides a struct with fields that describe the details of logger.
type Logger struct {
	out       io.Writer
	fmtter    Formatter
	lvl       Level
	callDepth int
	skipLine  bool
	mu        sync.Mutex
}

// New returns a customized Logger
func New(out io.Writer, fmtter Formatter, lvl Level) *Logger {
	return &Logger{
		out:       out,
		fmtter:    fmtter,
		lvl:       lvl,
		callDepth: 3,
	}
}

// DefaultLogger returns the default logger within the pkg, i.e. the one used in log.Infof....
func DefaultLogger() *Logger {
	return logger
}

// SetOutput sets the output of Logger l
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.out = out
}

// SetFormatter sets the formatter of Logger l
func (l *Logger) SetFormatter(fmtter Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.fmtter = fmtter
}

// SetLevel sets the level of Logger l
func (l *Logger) SetLevel(lvl Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lvl = lvl
}

// SetOutput sets the output of default Logger
func SetOutput(out io.Writer) {
	logger.SetOutput(out)
}

// SetFormatter sets the formatter of default Logger
func SetFormatter(fmtter Formatter) {
	logger.SetFormatter(fmtter)
}

// SetLevel sets the level of default Logger
func SetLevel(lvl Level) {
	logger.SetLevel(lvl)
}

func (l *Logger) output(record *Record) (err error) {
	b, err := l.fmtter.Format(record)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, err = l.out.Write(b)

	return
}

// Debug ...
func (l *Logger) Debug(v ...interface{}) {
	if l.lvl <= DebugLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), DebugLevel)
		l.output(record)
	}
}

// Debugf ...
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.lvl <= DebugLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), DebugLevel)
		l.output(record)
	}
}

// Info ...
func (l *Logger) Info(v ...interface{}) {
	if l.lvl <= InfoLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), "", InfoLevel)
		l.output(record)
	}
}

// Infof ...
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.lvl <= InfoLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), "", InfoLevel)
		l.output(record)
	}
}

// Warning ...
func (l *Logger) Warning(v ...interface{}) {
	if l.lvl <= WarningLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), "", WarningLevel)
		l.output(record)
	}
}

// Warningf ...
func (l *Logger) Warningf(format string, v ...interface{}) {
	if l.lvl <= WarningLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), "", WarningLevel)
		l.output(record)
	}
}

// Error ...
func (l *Logger) Error(v ...interface{}) {
	if l.lvl <= ErrorLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), ErrorLevel)
		l.output(record)
	}
}

// Errorf ...
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.lvl <= ErrorLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), ErrorLevel)
		l.output(record)
	}
}

// Fatal ...
func (l *Logger) Fatal(v ...interface{}) {
	if l.lvl <= FatalLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), FatalLevel)
		l.output(record)
	}
	os.Exit(1)
}

// Fatalf ...
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.lvl <= FatalLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), FatalLevel)
		l.output(record)
	}
	os.Exit(1)
}

func (l *Logger) getLine() string {
	if l.skipLine {
		return ""
	}
	return line(l.callDepth)
}

// Debug ...
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Debugf ...
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Info ...
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Infof ...
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Warning  ...
func Warning(v ...interface{}) {
	logger.Warning(v...)
}

// Warningf ...
func Warningf(format string, v ...interface{}) {
	logger.Warningf(format, v...)
}

// Error ...
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Errorf ...
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Fatal ...
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Fatalf ...
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

func line(calldepth int) string {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}

	for i := len(file) - 2; i > 0; i-- {
		if file[i] == os.PathSeparator {
			file = file[i+1:]
			break
		}
	}

	return fmt.Sprintf("[%s:%d]:", file, line)
}
