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

package q

import (
	"reflect"
	"testing"
)

func TestMustClone(t *testing.T) {
	type args struct {
		query *Query
	}
	tests := []struct {
		name string
		args args
		want *Query
	}{
		{"ptr", args{New(KeyWords{"public": "true"})}, New(KeyWords{"public": "true"})},
		{"nil", args{nil}, New(KeyWords{})},
		{"sort", args{&Query{Keywords: KeyWords{"public": "true"}, Sorts: []*Sort{NewSort("col-1", true)}}}, &Query{Keywords: KeyWords{"public": "true"}, Sorts: []*Sort{NewSort("col-1", true)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustClone(tt.args.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustClone() = %v, want %v", got, tt.want)
			}
		})
	}
}
