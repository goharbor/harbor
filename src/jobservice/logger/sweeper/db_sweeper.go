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

package sweeper

import (
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/joblog"
)

var dbInit = make(chan int, 1)
var isDBInit = false

// DBSweeper is used to sweep the DB logs
type DBSweeper struct {
	duration int
}

// NewDBSweeper is constructor of DBSweeper
func NewDBSweeper(duration int) *DBSweeper {
	return &DBSweeper{
		duration: duration,
	}
}

// Sweep logs
func (dbs *DBSweeper) Sweep() (int, error) {
	// DB initialization not completed, waiting
	WaitingDBInit()

	// Start to sweep logs
	before := time.Now().Add(time.Duration(dbs.duration) * oneDay * -1)
	count, err := joblog.Mgr.DeleteBefore(orm.Context(), before)

	if err != nil {
		return 0, fmt.Errorf("sweep logs in DB failed before %s with error: %s", before, err)
	}

	return int(count), nil
}

// Duration for sweeping
func (dbs *DBSweeper) Duration() int {
	return dbs.duration
}

// WaitingDBInit waiting DB init
func WaitingDBInit() {
	if !isDBInit {
		<-dbInit
	}
}

// PrepareDBSweep invoked after DB init
func PrepareDBSweep() error {
	if !isDBInit {
		isDBInit = true
		dbInit <- 1
	}
	return nil
}
