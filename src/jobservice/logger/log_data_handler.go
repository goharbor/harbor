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
	"errors"

	"github.com/goharbor/harbor/src/jobservice/logger/getter"
)

// Retrieve is wrapper func for getter.Retrieve
func Retrieve(logID string) ([]byte, error) {
	val, ok := singletons.Load(systemKeyLogDataGetter)
	if !ok {
		return nil, errors.New("no log data getter is configured")
	}

	return val.(getter.Interface).Retrieve(logID)
}
