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
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/cache/memory"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	commitMain = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	commitV1   = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	commitNew  = "cccccccccccccccccccccccccccccccccccccccc"
	repoTiny   = "org/tiny-model"
)

// bigWeight is above the default 10 MiB replication chunk size so copy-by-chunk really chunks.
var bigWeight = bytes.Repeat([]byte("0123456789abcdef"), (10<<20)/16+4096)

func tinyModel() *hubtest.Model {
	files := func(extra string) map[string]hubtest.File {
		return map[string]hubtest.File{
			".gitattributes":    {Content: []byte("*.safetensors filter=lfs\n")},
			"README.md":         {Content: []byte("# Tiny " + extra + "\n")},
			"config.json":       {Content: []byte(`{"model_type":"tiny"}`)},
			"model.safetensors": {Content: bigWeight, LFS: true},
			"sub/modeling.py":   {Content: []byte("print('hi')\n")},
		}
	}
	return &hubtest.Model{
		ID: "Org/Tiny-Model",
		Commits: map[string]*hubtest.Commit{
			commitMain: {Files: files("main"), LastModified: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), License: "mit"},
			commitV1:   {Files: files("v1"), LastModified: time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC), License: "mit"},
			commitNew:  {Files: files("new"), LastModified: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC), License: "mit"},
		},
		Branches: map[string]string{"main": commitMain, "feature/x": commitMain},
		Tags:     map[string]string{"v1": commitV1},
	}
}

func newMemoryCache() cache.Cache {
	c, _ := memory.New(cache.Options{Codec: cache.DefaultCodec()})
	return c
}

func hubRegistry(h *hubtest.Hub, id int64, token string) *model.Registry {
	r := &model.Registry{ID: id, Name: "hf", Type: model.RegistryTypeHuggingFace, URL: h.URL}
	if token != "" {
		r.Credential = &model.Credential{Type: model.CredentialTypeBasic, AccessKey: "any", AccessSecret: token}
	}
	return r
}

func newTestAdapter(t *testing.T, r *model.Registry, c cache.Cache) *adapter {
	t.Helper()
	a, err := newAdapter(r, c)
	require.NoError(t, err)
	return a
}
