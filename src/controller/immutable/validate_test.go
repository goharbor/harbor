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

package immutable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

// fakeManager records the writes of the controller without a database.
type fakeManager struct {
	stored  *model.Metadata
	created int
	updated int
	toggled int
}

func (f *fakeManager) CreateImmutableRule(context.Context, *model.Metadata) (int64, error) {
	f.created++
	return 1, nil
}

func (f *fakeManager) UpdateImmutableRule(context.Context, int64, *model.Metadata) error {
	f.updated++
	return nil
}

func (f *fakeManager) EnableImmutableRule(context.Context, int64, bool) error {
	f.toggled++
	return nil
}

func (f *fakeManager) GetImmutableRule(context.Context, int64) (*model.Metadata, error) {
	return f.stored, nil
}

func (f *fakeManager) Count(context.Context, *q.Query) (int64, error) { return 0, nil }

func (f *fakeManager) ListImmutableRules(context.Context, *q.Query) ([]*model.Metadata, error) {
	return nil, nil
}

func (f *fakeManager) DeleteImmutableRule(context.Context, int64) error { return nil }

func newRule(kind, decoration, tagPattern, repoPattern string) *model.Metadata {
	return &model.Metadata{
		ID:        1,
		ProjectID: 1,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{Kind: kind, Decoration: decoration, Pattern: tagPattern},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{Kind: "doublestar", Decoration: "repoMatches", Pattern: repoPattern},
			},
		},
	}
}

func TestCreateImmutableRuleRejectsRuleTheMatcherCannotEvaluate(t *testing.T) {
	cases := []struct {
		name string
		rule *model.Metadata
	}{
		{"unclosed character class in tag pattern", newRule("doublestar", "matches", "[release", "**")},
		{"unclosed alternation in tag pattern", newRule("doublestar", "matches", "{a,b", "**")},
		{"unclosed alternation in repository pattern", newRule("doublestar", "matches", "release-*", "{a,b")},
		{"unregistered selector kind", newRule("regex", "matches", "release-.*", "**")},
		{"unsupported decoration", newRule("doublestar", "equals", "release-1", "**")},
		{"empty kind and decoration", newRule("", "", "release-*", "**")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mgr := &fakeManager{}
			ctr := NewAPIController(mgr)

			_, err := ctr.CreateImmutableRule(context.Background(), c.rule)

			require.Error(t, err)
			assert.True(t, errors.IsErr(err, errors.BadRequestCode), "got %v", err)
			assert.Zero(t, mgr.created, "the invalid rule must not be stored")
		})
	}
}

func TestCreateImmutableRuleAcceptsValidRule(t *testing.T) {
	mgr := &fakeManager{}
	ctr := NewAPIController(mgr)
	rule := newRule("doublestar", "matches", "release-**", "{redis,mysql}")
	rule.TagSelectors = append(rule.TagSelectors, &model.Selector{Kind: "doublestar", Decoration: "excludes", Pattern: "release-[0-9]"})
	// nil selectors are skipped by the matcher, so they are not an error
	rule.TagSelectors = append(rule.TagSelectors, nil)

	id, err := ctr.CreateImmutableRule(context.Background(), rule)

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.Equal(t, 1, mgr.created)
}

func TestUpdateImmutableRuleRejectsRuleTheMatcherCannotEvaluate(t *testing.T) {
	mgr := &fakeManager{stored: newRule("doublestar", "matches", "release-**", "**")}
	ctr := NewAPIController(mgr)

	err := ctr.UpdateImmutableRule(context.Background(), 1, newRule("doublestar", "matches", "[release", "**"))

	require.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.BadRequestCode), "got %v", err)
	assert.Zero(t, mgr.updated, "the stored rule must stay as it was")
}

func TestUpdateImmutableRuleStillTogglesAnAlreadyInvalidRule(t *testing.T) {
	// a rule stored before the validation existed can still be disabled
	mgr := &fakeManager{stored: newRule("doublestar", "matches", "[release", "**")}
	ctr := NewAPIController(mgr)
	toggled := newRule("doublestar", "matches", "[release", "**")
	toggled.Disabled = true

	err := ctr.UpdateImmutableRule(context.Background(), 1, toggled)

	require.NoError(t, err)
	assert.Equal(t, 1, mgr.toggled)
	assert.Zero(t, mgr.updated)
}
