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

package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchCatalogURLPattern(t *testing.T) {
	cases := []struct {
		url   string
		match bool
	}{
		{
			url:   "/v2/_catalog",
			match: true,
		},
		{
			url:   "/v2/_catalog/",
			match: true,
		},
		{
			url:   "/v2/_catalog/xxx",
			match: false,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.match, len(V2CatalogURLRe.FindStringSubmatch(c.url)) == 1)
	}
}
