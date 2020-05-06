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
	"bytes"
	"os"
	"strings"
	"testing"
)

var (
	message = "message"
)

// contains reports whether the string is contained in the log.
func contains(t *testing.T, str string, lvl string, line, msg string) bool {
	return strings.Contains(str, lvl) && strings.Contains(str, line) && strings.Contains(str, msg)
}

func TestSetx(t *testing.T) {
	logger := New(nil, nil, WarningLevel)
	logger.setOutput(os.Stdout)
	fmt := NewTextFormatter()
	logger.setFormatter(fmt)
	logger.setLevel(DebugLevel)

	if logger.out != os.Stdout {
		t.Errorf("unexpected outer: %v != %v", logger.out, os.Stdout)
	}

	if logger.fmtter != fmt {
		t.Errorf("unexpected formatter: %v != %v", logger.fmtter, fmt)
	}

	if logger.lvl != DebugLevel {
		t.Errorf("unexpected log level: %v != %v", logger.lvl, DebugLevel)
	}
}

func TestWithFields(t *testing.T) {
	buf := enter()
	defer exit()

	logger.WithFields(Fields{"action": "create"}).Info(message)

	str := buf.String()

	var (
		expectedLevel = InfoLevel.string()
		expectLine    = `[action="create"]`
		expectMsg     = "message"
	)

	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestDebug(t *testing.T) {
	buf := enter()
	defer exit()

	Debug(message)

	str := buf.String()
	if str != "" {
		t.Errorf("unexpected message: %s != %s", str, "")
	}
}

func TestDebugf(t *testing.T) {
	buf := enter()
	defer exit()

	Debugf("%s", message)

	str := buf.String()
	if str != "" {
		t.Errorf("unexpected message: %s != %s", str, "")
	}
}

func TestInfo(t *testing.T) {
	var (
		expectedLevel = InfoLevel.string()
		expectLine    = ""
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	Info(message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestInfof(t *testing.T) {
	var (
		expectedLevel = InfoLevel.string()
		expectLine    = ""
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	Infof("%s", message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestWarning(t *testing.T) {
	var (
		expectedLevel = WarningLevel.string()
		expectLine    = ""
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	Warning(message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestWarningf(t *testing.T) {
	var (
		expectedLevel = WarningLevel.string()
		expectLine    = ""
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	Warningf("%s", message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestError(t *testing.T) {
	var (
		expectedLevel = ErrorLevel.string()
		expectLine    = "logger_test.go:178"
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	Error(message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestErrorf(t *testing.T) {
	var (
		expectedLevel = ErrorLevel.string()
		expectLine    = "logger_test.go:196"
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	Errorf("%s", message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func TestDefaultLoggerErrorf(t *testing.T) {
	var (
		expectedLevel = ErrorLevel.string()
		expectLine    = "logger_test.go:214"
		expectMsg     = "message"
	)

	buf := enter()
	defer exit()

	DefaultLogger().Errorf("%s", message)

	str := buf.String()
	if !contains(t, str, expectedLevel, expectLine, expectMsg) {
		t.Errorf("unexpected message: %s, expected level: %s, expected line: %s, expected message: %s", str, expectedLevel, expectLine, expectMsg)
	}
}

func enter() *bytes.Buffer {
	b := make([]byte, 0, 32)
	buf := bytes.NewBuffer(b)

	logger.setOutput(buf)

	return buf
}

func exit() {
	logger.setOutput(os.Stdout)
}
