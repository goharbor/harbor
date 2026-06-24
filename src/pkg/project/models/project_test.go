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

package models

import (
	"context"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
)

func TestIsAuthOnly(t *testing.T) {
	assert.Equal(t, "auth_only", ProjectAuthOnly)

	p := &Project{Metadata: map[string]string{"public": "auth_only"}}
	assert.True(t, p.IsAuthOnly())
	assert.False(t, p.IsPublic())

	p = &Project{Metadata: map[string]string{"public": "true"}}
	assert.False(t, p.IsAuthOnly())
	assert.True(t, p.IsPublic())

	p = &Project{Metadata: map[string]string{"public": "false"}}
	assert.False(t, p.IsAuthOnly())
	assert.False(t, p.IsPublic())

	p = &Project{Metadata: map[string]string{}}
	assert.False(t, p.IsAuthOnly())

	p = &Project{}
	assert.False(t, p.IsAuthOnly())
}

// captureQS captures the SQL expression passed to FilterRaw so we can assert
// that FilterByPublic generates the correct subquery for each access level.
type captureQS struct {
	orm.QuerySeter
	expr string
}

func (c *captureQS) FilterRaw(_ string, expr string) orm.QuerySeter {
	c.expr = expr
	return c
}

func TestFilterByPublic(t *testing.T) {
	p := &Project{}

	cases := []struct {
		name        string
		value       any
		wantContains string
	}{
		{"true bool selects public", true, "value = 'true'"},
		{"true string selects public", "true", "value = 'true'"},
		{"false bool selects non-public", false, "value != 'true'"},
		{"false string selects non-public", "false", "value != 'true'"},
		{"auth_only string selects auth_only", "auth_only", "value = 'auth_only'"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qs := &captureQS{}
			p.FilterByPublic(context.Background(), qs, "public", tc.value)
			assert.True(t, strings.Contains(qs.expr, tc.wantContains),
				"expected expr %q to contain %q", qs.expr, tc.wantContains)
		})
	}
}
