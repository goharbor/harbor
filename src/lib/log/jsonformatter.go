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
	"encoding/json"
	"fmt"
	"time"
)

// JsonFormatter represents a kind of formatter that formats the logs as plain text
type JsonFormatter struct {
	// timeFormat string
}

// NewTextFormatter returns a TextFormatter, the format of time is time.RFC3339
func NewJsonFormatter() Formatter {
	return &JsonFormatter{
		// timeFormat: defaultTimeFormat,
	}
}

// Format formats the logs as "time [level] line message"
func (t *JsonFormatter) Format(r *Record) (b []byte, err error) {
	//	s := fmt.Sprintf("%s [%s] ", r.Time.Format(t.timeFormat), r.Lvl.string())
	b, err = json.Marshal(r)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

// // SetTimeFormat sets time format of JsonFormatter if the parameter fmt is not null
// func (t *JsonFormatter) SetTimeFormat(fmt string) {
// 	if len(fmt) != 0 {
// 		t.timeFormat = fmt
// 	}
// }

func (c RecordTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(c).Format(time.RFC3339) + `"`), nil
}

func (l Level) MarshalJSON() (b []byte, err error) {
	return []byte(`"` + l.string() + `"`), nil
}
