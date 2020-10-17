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

package lib

import (
	"testing"
)

func TestToBool(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil", args{nil}, false},
		{"bool true", args{true}, true},
		{"bool false", args{false}, false},
		{"string true", args{"true"}, true},
		{"string True", args{"True"}, true},
		{"string 1", args{"1"}, true},
		{"string false", args{"false"}, false},
		{"string False", args{"False"}, false},
		{"string 0", args{"0"}, false},
		{"int 1", args{1}, true},
		{"int 0", args{0}, false},
		{"int64 1", args{int64(1)}, true},
		{"int64 0", args{int64(0)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToBool(tt.args.v); got != tt.want {
				t.Errorf("ToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}
