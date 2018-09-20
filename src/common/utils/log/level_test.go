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
	"testing"
)

func TestString(t *testing.T) {
	m := map[Level]string{
		DebugLevel:   "DEBUG",
		InfoLevel:    "INFO",
		WarningLevel: "WARNING",
		ErrorLevel:   "ERROR",
		FatalLevel:   "FATAL",
		-1:           "UNKNOWN",
	}

	for level, str := range m {
		if level.string() != str {
			t.Errorf("unexpected string: %s != %s", level.string(), str)
		}
	}
}

func TestParseLevel(t *testing.T) {
	m := map[string]Level{
		"DEBUG":   DebugLevel,
		"INFO":    InfoLevel,
		"WARNING": WarningLevel,
		"ERROR":   ErrorLevel,
		"FATAL":   FatalLevel,
	}

	for str, level := range m {
		l, err := parseLevel(str)
		if err != nil {
			t.Errorf("failed to parse level: %v", err)
		}
		if l != level {
			t.Errorf("unexpected level: %d != %d", l, level)
		}
	}

	if _, err := parseLevel("UNKNOWN"); err == nil {
		t.Errorf("unexpected behaviour: should be error here")
	}
}
