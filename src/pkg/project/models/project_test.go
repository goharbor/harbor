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
		name         string
		value        any
		wantContains string
	}{
		{"true bool selects public", true, "value = 'true'"},
		{"true string selects public", "true", "value = 'true'"},
		{"false bool selects private", false, "value = 'false'"},
		{"false string selects private", "false", "value = 'false'"},
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

func TestFilterByMember(t *testing.T) {
	p := &Project{}

	cases := []struct {
		name           string
		query          *MemberQuery
		wantContains   []string
		wantNotContain string
	}{
		{
			name:           "member only",
			query:          &MemberQuery{UserID: 1},
			wantContains:   []string{"entity_id = 1"},
			wantNotContain: "auth_only",
		},
		{
			name:         "with public",
			query:        &MemberQuery{UserID: 1, WithPublic: true},
			wantContains: []string{"entity_id = 1", "value = 'true'"},
		},
		{
			name:         "with auth_only",
			query:        &MemberQuery{UserID: 1, WithAuthOnly: true},
			wantContains: []string{"entity_id = 1", "value = 'auth_only'"},
		},
		{
			name:         "with public and auth_only",
			query:        &MemberQuery{UserID: 1, WithPublic: true, WithAuthOnly: true},
			wantContains: []string{"entity_id = 1", "value = 'true'", "value = 'auth_only'"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qs := &captureQS{}
			p.FilterByMember(context.Background(), qs, "member", tc.query)
			for _, want := range tc.wantContains {
				assert.True(t, strings.Contains(qs.expr, want),
					"expected expr %q to contain %q", qs.expr, want)
			}
			if tc.wantNotContain != "" {
				assert.False(t, strings.Contains(qs.expr, tc.wantNotContain),
					"expected expr %q to not contain %q", qs.expr, tc.wantNotContain)
			}
		})
	}

	t.Run("non-MemberQuery value is a no-op", func(t *testing.T) {
		qs := &captureQS{}
		result := p.FilterByMember(context.Background(), qs, "member", "not-a-query")
		assert.Equal(t, qs, result)
		assert.Empty(t, qs.expr)
	})
}

func TestFilterByNames(t *testing.T) {
	p := &Project{}

	cases := []struct {
		name         string
		query        *NamesQuery
		wantContains []string
	}{
		{
			name:         "names only",
			query:        &NamesQuery{Names: []string{"library"}},
			wantContains: []string{"name IN ('library')"},
		},
		{
			name:         "with public",
			query:        &NamesQuery{Names: []string{"library"}, WithPublic: true},
			wantContains: []string{"name IN ('library')", "value = 'true'"},
		},
		{
			name:         "with auth_only",
			query:        &NamesQuery{Names: []string{"library"}, WithAuthOnly: true},
			wantContains: []string{"name IN ('library')", "value = 'auth_only'"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			qs := &captureQS{}
			p.FilterByNames(context.Background(), qs, "names", tc.query)
			for _, want := range tc.wantContains {
				assert.True(t, strings.Contains(qs.expr, want),
					"expected expr %q to contain %q", qs.expr, want)
			}
		})
	}

	t.Run("empty names is a no-op", func(t *testing.T) {
		qs := &captureQS{}
		result := p.FilterByNames(context.Background(), qs, "names", &NamesQuery{})
		assert.Equal(t, qs, result)
		assert.Empty(t, qs.expr)
	})
}
