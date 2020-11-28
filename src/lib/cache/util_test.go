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

package cache

import (
	"testing"
)

func Test_simpleCopy(t *testing.T) {
	var i int

	st := struct {
		x int
	}{}

	type args struct {
		dst interface{}
		src interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"copy to not addressable value", args{1, 1}, true},
		{"copy to not addressable value", args{i, 1}, true},
		{"copy addressable value", args{&i, 1}, false},
		{"copy to not convertible value", args{&st, 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := simpleCopy(tt.args.dst, tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("simpleCopy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
