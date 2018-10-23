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
	"fmt"
	"testing"
)

func TestServiceLogger(t *testing.T) {
	testingLogger := &fakeLogger{}
	SetLogger(testingLogger)

	Debug("DEBUG")
	Debugf("%s\n", "DEBUGF")
	Info("INFO")
	Infof("%s\n", "INFOF")
	Warning("WARNING")
	Warningf("%s\n", "WARNINGF")
	Error("ERROR")
	Errorf("%s\n", "ERRORF")
	Fatal("FATAL")
	Fatalf("%s\n", "FATALF")
}

type fakeLogger struct{}

// For debuging
func (fl *fakeLogger) Debug(v ...interface{}) {
	fmt.Println(v...)
}

// For debuging with format
func (fl *fakeLogger) Debugf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// For logging info
func (fl *fakeLogger) Info(v ...interface{}) {
	fmt.Println(v...)
}

// For logging info with format
func (fl *fakeLogger) Infof(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// For warning
func (fl *fakeLogger) Warning(v ...interface{}) {
	fmt.Println(v...)
}

// For warning with format
func (fl *fakeLogger) Warningf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// For logging error
func (fl *fakeLogger) Error(v ...interface{}) {
	fmt.Println(v...)
}

// For logging error with format
func (fl *fakeLogger) Errorf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// For fatal error
func (fl *fakeLogger) Fatal(v ...interface{}) {
	fmt.Println(v...)
}

// For fatal error with error
func (fl *fakeLogger) Fatalf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
