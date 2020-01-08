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

package orm

import (
	"errors"
	"github.com/astaxie/beego/orm"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"strings"
)

// IsNotFoundError checks whether the err is orm.ErrNoRows. If it it, wrap it
// as a src/internal/error.Error with not found error code
func IsNotFoundError(err error, messageFormat string, args ...interface{}) (*ierror.Error, bool) {
	if errors.Is(err, orm.ErrNoRows) {
		e := ierror.NotFoundError(err)
		if len(messageFormat) > 0 {
			e.WithMessage(messageFormat, args...)
		}
		return e, true
	}
	return nil, false
}

// IsConflictError checks whether the err is duplicate key error. If it it, wrap it
// as a src/internal/error.Error with conflict error code
func IsConflictError(err error, messageFormat string, args ...interface{}) (*ierror.Error, bool) {
	if err == nil {
		return nil, false
	}
	if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		e := ierror.ConflictError(err)
		if len(messageFormat) > 0 {
			e.WithMessage(messageFormat, args...)
		}
		return e, true
	}
	return nil, false
}
