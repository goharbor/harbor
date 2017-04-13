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
	"time"
)

// Record holds information about log
type Record struct {
	Time time.Time // time when the log produced
	Msg  string    // content of the log
	Line string    // in which file and line that the log produced
	Lvl  Level     // level of the log
}

// NewRecord creates a record according to the arguments provided and returns it
func NewRecord(time time.Time, msg, line string, lvl Level) *Record {
	return &Record{
		Time: time,
		Msg:  msg,
		Line: line,
		Lvl:  lvl,
	}
}
