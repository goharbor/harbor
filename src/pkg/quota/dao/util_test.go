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

package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/q"
)

func Test_listOrderBy(t *testing.T) {
	query := func(sort string) *q.Query {
		return &q.Query{
			Sorting: sort,
		}
	}

	type args struct {
		query *q.Query
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no query", args{nil}, "b.creation_time DESC"},
		{"order by unsupported field", args{query("unknown")}, "b.creation_time DESC"},
		{"order by storage of hard", args{query("hard.storage")}, "(CAST( (CASE WHEN (hard->>'storage') IS NULL THEN '0' WHEN (hard->>'storage') = '-1' THEN '9223372036854775807' ELSE (hard->>'storage') END) AS BIGINT )) ASC"},
		{"order by unsupported hard resource", args{query("hard.unknown")}, "b.creation_time DESC"},
		{"order by storage of used", args{query("used.storage")}, "(CAST( (CASE WHEN (used->>'storage') IS NULL THEN '0' WHEN (used->>'storage') = '-1' THEN '9223372036854775807' ELSE (used->>'storage') END) AS BIGINT )) ASC"},
		{"order by unsupported used resource", args{query("used.unknown")}, "b.creation_time DESC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listOrderBy(tt.args.query); got != tt.want {
				t.Errorf("listOrderBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
