//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/goharbor/harbor/src/lib/config"
)

func TestGetGCMaxWorkers(t *testing.T) {
	// unset means no limit
	assert.Equal(t, 0, GetGCMaxWorkers())

	cases := []struct {
		env      string
		expected int
	}{
		{"8", 8},
		{"1", 1},
		{"100", 100},
		// invalid values fall back to no limit rather than blocking GC
		{"0", 0},
		{"-1", 0},
		{"abc", 0},
		{"", 0},
		{"1.5", 0},
	}
	for _, c := range cases {
		t.Run(c.env, func(t *testing.T) {
			t.Setenv("GC_MAX_WORKERS", c.env)
			assert.Equal(t, c.expected, GetGCMaxWorkers())
		})
	}
}
