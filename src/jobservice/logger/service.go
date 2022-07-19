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

package logger

import (
	"github.com/goharbor/harbor/src/lib/log"
)

// jobServiceLogger is used to log for job service itself
func jobServiceLogger() (Interface, bool) {
	val, ok := singletons.Load(systemKeyServiceLogger)
	if ok {
		return val.(Interface), ok
	}

	return nil, false
}

// Debug ...
func Debug(v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Debug(v...)
	} else {
		log.Debug(v...)
	}
}

// Debugf for debuging with format
func Debugf(format string, v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Debugf(format, v...)
	} else {
		log.Debugf(format, v...)
	}
}

// Info ...
func Info(v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Info(v...)
	} else {
		log.Info(v...)
	}
}

// Infof for logging info with format
func Infof(format string, v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Infof(format, v...)
	} else {
		log.Infof(format, v...)
	}
}

// Warning ...
func Warning(v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Warning(v...)
	} else {
		log.Warning(v...)
	}
}

// Warningf for warning with format
func Warningf(format string, v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Warningf(format, v...)
	} else {
		log.Warningf(format, v...)
	}
}

// Error for logging error
func Error(v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Error(v...)
	} else {
		log.Error(v...)
	}
}

// Errorf for logging error with format
func Errorf(format string, v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Errorf(format, v...)
	} else {
		log.Errorf(format, v...)
	}
}

// Fatal ...
func Fatal(v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Fatal(v...)
	} else {
		log.Fatal(v...)
	}
}

// Fatalf for fatal error with error
func Fatalf(format string, v ...interface{}) {
	if jLogger, ok := jobServiceLogger(); ok {
		jLogger.Fatalf(format, v...)
	} else {
		log.Fatalf(format, v...)
	}
}
