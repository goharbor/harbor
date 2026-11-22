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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/q"
)

func TestPaginationOnRawSQL(t *testing.T) {
	// Standard pagination
	query := &q.Query{
		PageNumber: 1,
		PageSize:   10,
	}
	sql := "select * from my_table where id > ? order by id "
	params := []any{2}
	sql, params = PaginationOnRawSQL(query, sql, params)
	assert.Equal(t, "select * from my_table where id > ? order by id  limit ? offset ?", sql)
	assert.Equal(t, int64(10), params[1])
	assert.Equal(t, int64(0), params[2])

	// Pagination with ExtraLimit (e.g. +1 for page-size-plus-one)
	queryWithExtra := &q.Query{
		PageNumber: 2,
		PageSize:   10,
		ExtraLimit: 1,
	}
	sql2 := "select * from my_table where id > ? order by id "
	params2 := []any{2}
	sql2, params2 = PaginationOnRawSQL(queryWithExtra, sql2, params2)
	assert.Equal(t, "select * from my_table where id > ? order by id  limit ? offset ?", sql2)
	assert.Equal(t, int64(11), params2[1]) // limit is PageSize + ExtraLimit (10 + 1 = 11)
	assert.Equal(t, int64(10), params2[2]) // offset is still (PageNumber - 1) * PageSize (1 * 10 = 10)

	// Pagination with negative ExtraLimit that exceeds PageSize clamps to 0
	queryWithNegativeExtra := &q.Query{
		PageNumber: 1,
		PageSize:   10,
		ExtraLimit: -20,
	}
	sql3 := "select * from my_table where id > ? order by id "
	params3 := []any{2}
	sql3, params3 = PaginationOnRawSQL(queryWithNegativeExtra, sql3, params3)
	assert.Equal(t, "select * from my_table where id > ? order by id  limit ? offset ?", sql3)
	assert.Equal(t, int64(0), params3[1]) // clamped to 0
}
