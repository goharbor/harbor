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
	"os"
	"time"

	"github.com/astaxie/beego/logs"
)

// BeeLevel2Hbr mapping beego's log level to harbor's log lever
var BeeLevel2Hbr = map[int]Level{
	logs.LevelEmergency:     FatalLevel,
	logs.LevelAlert:         FatalLevel,
	logs.LevelCritical:      FatalLevel,
	logs.LevelError:         ErrorLevel,
	logs.LevelWarning:       WarningLevel,
	logs.LevelNotice:        InfoLevel,
	logs.LevelInformational: InfoLevel,
	logs.LevelDebug:         DebugLevel,
}

// NewHarborAdapter Create a logger adapter of Harbor
func NewHarborAdapter() logs.Logger {
	level, err := parseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		panic(err)
	}
	return &HbrBeeLogger{level: level}
}

// HbrBeeLogger is a adapter
type HbrBeeLogger struct {
	level Level
}

// Init implementing method. empty.
func (l *HbrBeeLogger) Init(jsonConfig string) error {
	return nil
}

// WriteMsg implementing method.
func (l *HbrBeeLogger) WriteMsg(when time.Time, msg string, level int) error {
	lvl, ok := BeeLevel2Hbr[level]
	if !ok {
		return fmt.Errorf("log level %d beego passed is undefined", level)
	}
	if lvl < l.level {
		return nil
	}

	// format logs to harbor style
	s := fmt.Sprintf("%s [%s] %s", when.Format(defaultTimeFormat), lvl.string(), msg)
	logger.writeBeeLog(s)
	return nil
}

// Destroy implementing method. empty.
func (l *HbrBeeLogger) Destroy() {
}

// Flush implementing method. empty.
func (l *HbrBeeLogger) Flush() {
}

// writeBeeLog write log messages to logger
func (l *Logger) writeBeeLog(msg string) {
	b := []byte(msg)
	if len(b) == 0 || b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.out.Write(b)
}
