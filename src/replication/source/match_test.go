// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		pattern string
		str     string
		matched bool
	}{
		{"", "", true},
		{"*", "library", true},
		{"library/*", "library/mysql", true},
		{"library/*", "library/mysql/5.6", false},
		{"library/mysq?", "library/mysql", true},
		{"library/mysq?", "library/mysqld", false},
	}

	for _, c := range cases {
		matched, err := match(c.pattern, c.str)
		require.Nil(t, err)
		assert.Equal(t, c.matched, matched)
	}
}
