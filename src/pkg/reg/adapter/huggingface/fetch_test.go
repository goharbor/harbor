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

package huggingface

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func fetchHub() *hubtest.Hub {
	other := &hubtest.Model{
		ID:       "Org/Other",
		Commits:  map[string]*hubtest.Commit{commitNew: {}},
		Branches: map[string]string{"main": commitNew},
	}
	invalid := &hubtest.Model{
		ID:       "Org/bad_-name",
		Commits:  map[string]*hubtest.Commit{commitNew: {}},
		Branches: map[string]string{"main": commitNew},
	}
	foreign := &hubtest.Model{
		ID:       "Else/Model",
		Commits:  map[string]*hubtest.Commit{commitNew: {}},
		Branches: map[string]string{"main": commitNew},
	}
	return hubtest.New(tinyModel(), other, invalid, foreign)
}

type fetched map[string][][]string

func summarize(resources []*model.Resource) fetched {
	out := fetched{}
	for _, r := range resources {
		for _, a := range r.Metadata.Artifacts {
			out[r.Metadata.Repository.Name] = append(out[r.Metadata.Repository.Name], a.Tags)
		}
	}
	return out
}

func TestFetchArtifacts(t *testing.T) {
	h := fetchHub()
	defer h.Close()
	reg := hubRegistry(h, 1, "")
	a := newTestAdapter(t, reg, newMemoryCache())

	name := func(v string) *model.Filter { return &model.Filter{Type: model.FilterTypeName, Value: v} }
	tag := func(v, decoration string) *model.Filter {
		return &model.Filter{Type: model.FilterTypeTag, Value: v, Decoration: decoration}
	}
	cases := []struct {
		name    string
		filters []*model.Filter
		want    fetched
	}{
		{
			name:    "author wildcard, case-insensitive pattern, invalid names skipped",
			filters: []*model.Filter{name("ORG/**")},
			want: fetched{
				"org/other":      {{"main"}},
				"org/tiny-model": {{"feature-x", "main"}, {"v1"}},
			},
		},
		{
			name:    "one artifact per commit carrying all its tags",
			filters: []*model.Filter{name("org/tiny-model")},
			want:    fetched{"org/tiny-model": {{"feature-x", "main"}, {"v1"}}},
		},
		{
			name:    "tag filter matches sanitized tags",
			filters: []*model.Filter{name("org/tiny-model"), tag("feature-*", model.Matches)},
			want:    fetched{"org/tiny-model": {{"feature-x"}}},
		},
		{
			name:    "tag filter excludes",
			filters: []*model.Filter{name("org/tiny-model"), tag("main", model.Excludes)},
			want:    fetched{"org/tiny-model": {{"feature-x"}, {"v1"}}},
		},
		{
			name:    "pattern over names",
			filters: []*model.Filter{name("org/tiny*")},
			want:    fetched{"org/tiny-model": {{"feature-x", "main"}, {"v1"}}},
		},
		{
			name:    "several authors",
			filters: []*model.Filter{name("{org,else}/*"), tag("main", model.Matches)},
			want: fetched{
				"else/model":     {{"main"}},
				"org/other":      {{"main"}},
				"org/tiny-model": {{"main"}},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resources, err := a.FetchArtifacts(c.filters)
			require.NoError(t, err)
			assert.Equal(t, c.want, summarize(resources))
			for _, r := range resources {
				assert.Equal(t, model.ResourceTypeArtifact, r.Type)
				assert.Equal(t, reg, r.Registry)
			}
		})
	}
}

func TestFetchArtifactsRequiresAuthor(t *testing.T) {
	a := newTestAdapter(t, &model.Registry{}, newMemoryCache())
	for _, filters := range [][]*model.Filter{
		nil,
		{{Type: model.FilterTypeTag, Value: "main"}},
		{{Type: model.FilterTypeName, Value: "**"}},
		{{Type: model.FilterTypeName, Value: "*/model"}},
		{{Type: model.FilterTypeName, Value: "gpt2"}},
		{{Type: model.FilterTypeName, Value: 1}},
	} {
		_, err := a.FetchArtifacts(filters)
		assert.True(t, errors.IsErr(err, errors.BadRequestCode), err)
	}
}

func TestFetchArtifactsFailsWholeRun(t *testing.T) {
	h := fetchHub()
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	h.RateLimit("refs", 100)
	resources, err := a.FetchArtifacts([]*model.Filter{{Type: model.FilterTypeName, Value: "org/**"}})
	assert.True(t, errors.IsRateLimitError(err), err)
	assert.Nil(t, resources)

	// A pinned model that does not exist fails the run as not found, although the anonymous Hub
	// answers 401 for it.
	h.RateLimit("refs", 0)
	_, err = a.FetchArtifacts([]*model.Filter{{Type: model.FilterTypeName, Value: "org/missing"}})
	assert.True(t, errors.IsNotFoundErr(err), err)
}

func TestFetchArtifactsAPIBudget(t *testing.T) {
	h := fetchHub()
	h.PageSize = 1000
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	_, err := a.FetchArtifacts([]*model.Filter{{Type: model.FilterTypeName, Value: "org/**"}})
	require.NoError(t, err)
	assert.Equal(t, 1, h.Calls("account"))
	assert.Equal(t, 1, h.Calls("list"))
	assert.Equal(t, 2, h.Calls("refs"))
	assert.Equal(t, 0, h.Calls("revision"))
	assert.Equal(t, 0, h.Calls("resolve"))
}
