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
	"sort"
	"strings"
	"sync"
	"time"
)

// NOTE: the default depth for the logger is 3 so that we can get the correct file and line when use the logger to log message
var logger = New(os.Stdout, NewTextFormatter(), WarningLevel, 3)

const srcSeparator = "harbor" + string(os.PathSeparator) + "src"

func init() {
	lvl := os.Getenv("LOG_LEVEL")
	if len(lvl) == 0 {
		logger.setLevel(InfoLevel)
		return
	}

	level, err := parseLevel(lvl)
	if err != nil {
		logger.setLevel(InfoLevel)
		return
	}

	logger.setLevel(level)
}

// Fields type alias to map[string]any
type Fields = map[string]any

// Logger provides a struct with fields that describe the details of logger.
type Logger struct {
	out       io.Writer
	fmtter    Formatter
	lvl       Level
	callDepth int
	skipLine  bool
	fields    map[string]any
	fieldsStr string
	mu        *sync.Mutex // ptr here to share one sync.Mutex for clone method
	fallback  *Logger     // fallback logger when current out fail
}

// New returns a customized Logger
func New(out io.Writer, fmtter Formatter, lvl Level, options ...any) *Logger {
	// Default set to be 3
	depth := 3
	// If passed in as option, then reset depth
	// Use index 0
	if len(options) > 0 {
		d, ok := options[0].(int)
		if ok && d > 0 {
			depth = d
		}
	}

	return &Logger{
		out:       out,
		fmtter:    fmtter,
		lvl:       lvl,
		callDepth: depth,
		fields:    map[string]any{},
		mu:        &sync.Mutex{},
	}
}

// DefaultLogger returns the default logger within the pkg, i.e. the one used in log.Infof....
func DefaultLogger() *Logger {
	return logger
}

// SetFallback enable fallback when error happen
func (l *Logger) SetFallback(logger *Logger) {
	l.fallback = logger
}

func (l *Logger) clone() *Logger {
	return &Logger{
		out:       l.out,
		fmtter:    l.fmtter,
		lvl:       l.lvl,
		callDepth: l.callDepth,
		skipLine:  l.skipLine,
		fields:    l.fields,
		fieldsStr: l.fieldsStr,
		mu:        l.mu,
	}
}

// WithDepth returns cloned logger with new depth
func (l *Logger) WithDepth(depth int) *Logger {
	r := l.clone()
	r.callDepth = depth

	return r
}

// WithFields returns cloned logger which fields merged with the new fields
func (l *Logger) WithFields(fields Fields) *Logger {
	r := l.clone()

	if len(fields) > 0 {
		copyFields := make(map[string]any, len(l.fields)+len(fields))
		for key, value := range l.fields {
			copyFields[key] = value
		}
		for key, value := range fields {
			copyFields[key] = value
		}

		sortedKeys := make([]string, 0, len(copyFields))
		for key := range copyFields {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)

		parts := make([]string, 0, len(copyFields))
		for _, key := range sortedKeys {
			parts = append(parts, fmt.Sprintf(`%v="%v"`, key, copyFields[key]))
		}

		r.fields = copyFields
		r.fieldsStr = "[" + strings.Join(parts, " ") + "]"
	}

	return r
}

// WithField returns cloned logger which fields merged with field key=value
func (l *Logger) WithField(key string, value any) *Logger {
	return l.WithFields(Fields{key: value})
}

// SetOutput sets the output of Logger l
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.out = out
}

// setFormatter sets the formatter of Logger l
func (l *Logger) setFormatter(fmtter Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.fmtter = fmtter
}

// setLevel sets the level of Logger l
func (l *Logger) setLevel(lvl Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lvl = lvl
}

func (l *Logger) output(record *Record) (err error) {
	b, err := l.fmtter.Format(record)
	if err != nil {
		return
	}
	defer func() {
		if err := recover(); err != nil && l.fallback != nil {
			_ = l.fallback.output(record)
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = l.out.Write(b)
	if err != nil && l.fallback != nil {
		_ = l.fallback.output(record)
	}
	return
}

// Debug ...
func (l *Logger) Debug(v ...any) {
	if l.lvl <= DebugLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), DebugLevel)
		_ = l.output(record)
	}
}

// Debugf ...
func (l *Logger) Debugf(format string, v ...any) {
	if l.lvl <= DebugLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), DebugLevel)
		_ = l.output(record)
	}
}

// Info ...
func (l *Logger) Info(v ...any) {
	if l.lvl <= InfoLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), InfoLevel)
		_ = l.output(record)
	}
}

// Infof ...
func (l *Logger) Infof(format string, v ...any) {
	if l.lvl <= InfoLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), InfoLevel)
		_ = l.output(record)
	}
}

// Warning ...
func (l *Logger) Warning(v ...any) {
	if l.lvl <= WarningLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), WarningLevel)
		_ = l.output(record)
	}
}

// Warningf ...
func (l *Logger) Warningf(format string, v ...any) {
	if l.lvl <= WarningLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), WarningLevel)
		_ = l.output(record)
	}
}

// Error ...
func (l *Logger) Error(v ...any) {
	if l.lvl <= ErrorLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), ErrorLevel)
		_ = l.output(record)
	}
}

// Errorf ...
func (l *Logger) Errorf(format string, v ...any) {
	if l.lvl <= ErrorLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), ErrorLevel)
		_ = l.output(record)
	}
}

// Fatal ...
func (l *Logger) Fatal(v ...any) {
	if l.lvl <= FatalLevel {
		record := NewRecord(time.Now(), fmt.Sprint(v...), l.getLine(), FatalLevel)
		_ = l.output(record)
	}
	os.Exit(1)
}

// Fatalf ...
func (l *Logger) Fatalf(format string, v ...any) {
	if l.lvl <= FatalLevel {
		record := NewRecord(time.Now(), fmt.Sprintf(format, v...), l.getLine(), FatalLevel)
		_ = l.output(record)
	}
	os.Exit(1)
}

// GetLevel returns the verbosity level of this logger
func (l *Logger) GetLevel() Level {
	return l.lvl
}

func (l *Logger) getLine() string {
	var str string
	if !l.skipLine {
		str = line(l.callDepth)
	}

	str = str + l.fieldsStr

	if str != "" {
		str = str + ":"
	}

	return str
}

// Debug ...
func Debug(v ...any) {
	logger.WithDepth(4).Debug(v...)
}

// Debugf ...
func Debugf(format string, v ...any) {
	logger.WithDepth(4).Debugf(format, v...)
}

// Info ...
func Info(v ...any) {
	logger.WithDepth(4).Info(v...)
}

// Infof ...
func Infof(format string, v ...any) {
	logger.WithDepth(4).Infof(format, v...)
}

// Warning  ...
func Warning(v ...any) {
	logger.WithDepth(4).Warning(v...)
}

// Warningf ...
func Warningf(format string, v ...any) {
	logger.WithDepth(4).Warningf(format, v...)
}

// Error ...
func Error(v ...any) {
	logger.WithDepth(4).Error(v...)
}

// Errorf ...
func Errorf(format string, v ...any) {
	logger.WithDepth(4).Errorf(format, v...)
}

// Fatal ...
func Fatal(v ...any) {
	logger.WithDepth(4).Fatal(v...)
}

// Fatalf ...
func Fatalf(format string, v ...any) {
	logger.WithDepth(4).Fatalf(format, v...)
}

// GetLevel return the verbosity level of default logger
func GetLevel() Level {
	return logger.GetLevel()
}

func line(callDepth int) string {
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "???"

		line = 0
	}
	l := strings.SplitN(file, srcSeparator, 2)
	if len(l) > 1 {
		file = l[1]
	}
	return fmt.Sprintf("[%s:%d]", file, line)
}
