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

package getter

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/joblog"
)

// DBGetter is responsible for retrieving DB log data
type DBGetter struct {
}

// NewDBGetter is constructor of DBGetter
func NewDBGetter() *DBGetter {
	return &DBGetter{}
}

// Retrieve implements @Interface.Retrieve
func (dbg *DBGetter) Retrieve(logID string) ([]byte, error) {
	if len(logID) == 0 {
		return nil, errors.New("empty log identify")
	}

	jobLog, err := joblog.Mgr.Get(orm.Context(), logID)
	if err != nil {
		// Other errors have been ignored by GetJobLog()
		return nil, errs.NoObjectFoundError(fmt.Sprintf("log entity: %s", logID))
	}

	sz := int64(len(jobLog.Content))
	var buf []byte
	sizeLimit := logSizeLimit()
	if sizeLimit <= 0 || sz <= sizeLimit {
		buf = []byte(jobLog.Content)
		return buf, nil
	}
	if sz > sizeLimit {
		buf = []byte(jobLog.Content[sz-sizeLimit:])
	}
	return buf, nil
}
