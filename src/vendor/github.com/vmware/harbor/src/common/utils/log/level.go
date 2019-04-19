// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"strings"
)

// Level ...
type Level int

const (
	// DebugLevel debug
	DebugLevel Level = iota
	// InfoLevel info
	InfoLevel
	// WarningLevel warning
	WarningLevel
	// ErrorLevel error
	ErrorLevel
	// FatalLevel fatal
	FatalLevel
)

func (l Level) string() (lvl string) {
	switch l {
	case DebugLevel:
		lvl = "DEBUG"
	case InfoLevel:
		lvl = "INFO"
	case WarningLevel:
		lvl = "WARNING"
	case ErrorLevel:
		lvl = "ERROR"
	case FatalLevel:
		lvl = "FATAL"
	default:
		lvl = "UNKNOWN"
	}

	return
}

func parseLevel(lvl string) (level Level, err error) {

	switch strings.ToLower(lvl) {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warning":
		level = WarningLevel
	case "error":
		level = ErrorLevel
	case "fatal":
		level = FatalLevel
	default:
		err = fmt.Errorf("invalid log level: %s", lvl)
	}

	return
}
