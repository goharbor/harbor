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
	"github.com/goharbor/harbor/src/internal/error"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsNotFoundError(t *testing.T) {
	// nil error
	_, ok := IsNotFoundError(nil, "")
	assert.False(t, ok)

	// common error
	_, ok = IsNotFoundError(errors.New("common error"), "")
	assert.False(t, ok)

	// pass
	message := "message"
	e, ok := IsNotFoundError(orm.ErrNoRows, message)
	assert.True(t, ok)
	assert.Equal(t, error.NotFoundCode, e.Code)
	assert.Equal(t, message, e.Message)
}

func TestIsConflictError(t *testing.T) {
	// nil error
	_, ok := IsConflictError(nil, "")
	assert.False(t, ok)

	// common error
	_, ok = IsConflictError(errors.New("common error"), "")
	assert.False(t, ok)

	// pass
	message := "message"
	e, ok := IsConflictError(errors.New("duplicate key value violates unique constraint"), message)
	assert.True(t, ok)
	assert.Equal(t, error.ConflictCode, e.Code)
	assert.Equal(t, message, e.Message)
}
