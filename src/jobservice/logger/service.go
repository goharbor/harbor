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
	"log"
)

// jobServiceLogger is used to log for job service itself
var jobServiceLogger Interface

// Debug ...
func Debug(v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Debug(v...)
		return
	}

	log.Println(v...)
}

// Debugf for debuging with format
func Debugf(format string, v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Debugf(format, v...)
		return
	}

	log.Printf(format, v...)
}

// Info ...
func Info(v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Info(v...)
		return
	}

	log.Println(v...)
}

// Infof for logging info with format
func Infof(format string, v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Infof(format, v...)
		return
	}

	log.Printf(format, v...)
}

// Warning ...
func Warning(v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Warning(v...)
		return
	}

	log.Println(v...)
}

// Warningf for warning with format
func Warningf(format string, v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Warningf(format, v...)
		return
	}

	log.Printf(format, v...)
}

// Error for logging error
func Error(v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Error(v...)
		return
	}

	log.Println(v...)
}

// Errorf for logging error with format
func Errorf(format string, v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Errorf(format, v...)
		return
	}

	log.Printf(format, v...)
}

// Fatal ...
func Fatal(v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Fatal(v...)
		return
	}

	log.Fatal(v...)
}

// Fatalf for fatal error with error
func Fatalf(format string, v ...interface{}) {
	if jobServiceLogger != nil {
		jobServiceLogger.Fatalf(format, v...)
		return
	}

	log.Fatalf(format, v...)
}
