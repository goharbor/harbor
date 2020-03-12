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

package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_keyMatch2(t *testing.T) {
	type args struct {
		key1 string
		key2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match /project/1/robot, /project/1",
			args: args{"/project/1/robot", "/project/1"},
			want: false,
		},
		{
			name: "match /project/1/robot, /project/:pid",
			args: args{"/project/1/robot", "/project/:pid"},
			want: false,
		},
		{
			name: "match /project/1/robot, /project/1/*",
			args: args{"/project/1/robot", "/project/1/*"},
			want: true,
		},
		{
			name: "match /project/1/robot, /project/:pid/robot",
			args: args{"/project/1/robot", "/project/:pid/robot"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyMatch2(tt.args.key1, tt.args.key2); got != tt.want {
				t.Errorf("keyMatch2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexpStore(t *testing.T) {
	assert := assert.New(t)

	s := &regexpStore{}

	sLen := func() int {
		var l int
		s.entries.Range(func(key, value interface{}) bool {
			l++

			return true
		})
		return l
	}

	r1 := s.Get("key1", keyMatch2Build)
	r2 := s.Get("key1", keyMatch2Build)

	assert.Equal(r1, r2)
	s.Purge()
	assert.Equal(0, sLen())
}
