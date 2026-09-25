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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
)

func TestRepositoryOf(t *testing.T) {
	cases := []struct {
		modelID string
		want    string
		ok      bool
	}{
		{"Qwen/Qwen3-0.6B", "qwen/qwen3-0.6b", true},
		{"meta-llama/Llama-3.1-8B-Instruct", "meta-llama/llama-3.1-8b-instruct", true},
		{"org/model__name", "org/model__name", true},
		{"org/a_-b", "", false},
		{"org/a._b", "", false},
		{"org/name.", "", false},
		{"org/name-", "", false},
		{"gpt2", "", false},
		{"a/b/c", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		t.Run(c.modelID, func(t *testing.T) {
			got, ok := repositoryOf(c.modelID)
			assert.Equal(t, c.ok, ok)
			assert.Equal(t, c.want, got)
		})
	}
}

func TestTagOf(t *testing.T) {
	cases := []struct {
		ref  string
		want string
		ok   bool
	}{
		{"main", "main", true},
		{"v1.0", "v1.0", true},
		{"refs/pr/1", "refs-pr-1", true},
		{"feature/x y", "feature-x-y", true},
		{"ünicode", "-nicode", false},
		{"_ok", "_ok", true},
		{".hidden", "", false},
		{strings.Repeat("a", 128), strings.Repeat("a", 128), true},
		{strings.Repeat("a", 129), "", false},
		{"0123456789abcdef0123456789abcdef01234567", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		t.Run(c.ref, func(t *testing.T) {
			got, ok := tagOf(c.ref)
			assert.Equal(t, c.ok, ok)
			if ok {
				assert.Equal(t, c.want, got)
			}
		})
	}
}

func TestRefIndex(t *testing.T) {
	idx := newRefIndex([]hub.Ref{
		{Name: "main", Commit: commitMain},
		{Name: "feature/x", Commit: commitMain},
		{Name: "v1", Commit: commitV1},
		{Name: "a/b", Commit: commitV1},
		{Name: "a-b", Commit: commitNew},
		{Name: ".bad", Commit: commitNew},
	})
	assert.Equal(t, []string{"feature-x", "main", "v1"}, idx.tags())
	assert.Equal(t, map[string][]string{
		commitMain: {"feature-x", "main"},
		commitV1:   {"v1"},
	}, idx.tagsByCommit())

	c, err := idx.commit("feature-x")
	require.NoError(t, err)
	assert.Equal(t, commitMain, c)

	_, err = idx.commit("a-b")
	assert.True(t, errors.IsErr(err, errors.ConflictCode), err)
	assert.Contains(t, err.Error(), "a-b, a/b")

	_, err = idx.commit("missing")
	assert.True(t, errors.IsNotFoundErr(err), err)
}
