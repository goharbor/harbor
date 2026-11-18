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

package hub_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
)

const (
	commit1 = "1111111111111111111111111111111111111111"
	commit2 = "2222222222222222222222222222222222222222"
)

func fakeModel() *hubtest.Model {
	return &hubtest.Model{
		ID: "Org/Model-A",
		Commits: map[string]*hubtest.Commit{
			commit1: {
				LastModified: time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC),
				License:      []string{"mit", "apache-2.0"},
				Files: map[string]hubtest.File{
					"config.json":       {Content: []byte(`{"a":1}`)},
					"model.safetensors": {Content: []byte(strings.Repeat("w", 1000)), LFS: true},
					"dir/a b.txt":       {Content: []byte("spaces")},
				},
			},
		},
		Branches: map[string]string{"main": commit1},
		Tags:     map[string]string{"v1": commit1},
	}
}

func newClient(t *testing.T, h *hubtest.Hub, token string) *hub.Client {
	t.Helper()
	c, err := hub.New(h.URL, hub.Options{Token: token, Timeout: 10 * time.Second})
	require.NoError(t, err)
	return c
}

func TestDecodeSnapshotFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/qwen3-0.6b-revision.json")
	require.NoError(t, err)
	s, err := hub.DecodeSnapshot(data)
	require.NoError(t, err)
	assert.Equal(t, "Qwen/Qwen3-0.6B", s.ModelID)
	assert.Equal(t, "c1899de289a04d12100db370d81485cdf75e47ca", s.Commit)
	assert.Equal(t, time.Date(2025, 7, 26, 3, 46, 27, 0, time.UTC), s.LastModified)
	assert.Equal(t, []string{"apache-2.0"}, s.Licenses)
	require.Len(t, s.Files, 10)
	byPath := map[string]hub.File{}
	for _, f := range s.Files {
		byPath[f.Path] = f
	}
	st := byPath["model.safetensors"]
	assert.True(t, st.IsLFS())
	assert.Equal(t, "f47f71177f32bcd101b7573ec9171e6a57f4f4d31148d38e382306f42996874b", st.SHA256)
	assert.Equal(t, int64(1503300328), st.Size)
	cfg := byPath["config.json"]
	assert.False(t, cfg.IsLFS())
	assert.Equal(t, "f5c3703b78ae2a478ae15b247e9f855e0ce2107b", cfg.BlobID)
	assert.Equal(t, int64(726), cfg.Size)
}

func TestRefsFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/qwen3-0.6b-refs.json")
	require.NoError(t, err)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/models/Qwen/Qwen3-0.6B/refs", r.URL.Path)
		_, _ = w.Write(data)
	}))
	defer s.Close()
	c, err := hub.New(s.URL, hub.Options{})
	require.NoError(t, err)
	refs, err := c.Refs(context.Background(), "Qwen/Qwen3-0.6B")
	require.NoError(t, err)
	assert.Equal(t, []hub.Ref{{Name: "main", Commit: "c1899de289a04d12100db370d81485cdf75e47ca"}}, refs)
}

func TestDecodeSnapshotInvalid(t *testing.T) {
	cases := map[string]string{
		"no id":           `{"sha":"1111111111111111111111111111111111111111","siblings":[]}`,
		"bad sha":         `{"id":"a/b","sha":"main","siblings":[]}`,
		"no size":         `{"id":"a/b","sha":"1111111111111111111111111111111111111111","siblings":[{"rfilename":"x"}]}`,
		"bad lfs sha256":  `{"id":"a/b","sha":"1111111111111111111111111111111111111111","siblings":[{"rfilename":"x","size":1,"lfs":{"sha256":"zz","size":1}}]}`,
		"not json object": `[]`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := hub.DecodeSnapshot([]byte(body))
			assert.Error(t, err)
		})
	}
}

func TestDecodeSnapshotLicenses(t *testing.T) {
	cases := map[string][]string{
		`"mit"`:           {"mit"},
		`["mit","other"]`: {"mit", "other"},
		`""`:              nil,
		`null`:            nil,
		`{"x":1}`:         nil,
	}
	for license, want := range cases {
		s, err := hub.DecodeSnapshot([]byte(`{"id":"a/b","sha":"1111111111111111111111111111111111111111","cardData":{"license":` + license + `},"siblings":[]}`))
		require.NoError(t, err)
		assert.Equal(t, want, s.Licenses, license)
	}
}

func TestSnapshotAndRefs(t *testing.T) {
	h := hubtest.New(fakeModel())
	defer h.Close()
	c := newClient(t, h, "")
	ctx := context.Background()

	// The Hub is case-insensitive and redirects to the canonical model ID.
	s, err := c.Snapshot(ctx, "org/model-a", "main")
	require.NoError(t, err)
	assert.Equal(t, "Org/Model-A", s.ModelID)
	assert.Equal(t, commit1, s.Commit)
	assert.Equal(t, []string{"mit", "apache-2.0"}, s.Licenses)
	require.Len(t, s.Files, 3)
	assert.Equal(t, "dir/a b.txt", s.Files[1].Path)

	s, err = c.Snapshot(ctx, "org/model-a", commit1)
	require.NoError(t, err)
	assert.Equal(t, commit1, s.Commit)

	_, err = c.Snapshot(ctx, "org/model-a", commit2)
	assert.True(t, errors.IsNotFoundErr(err), err)

	refs, err := c.Refs(ctx, "org/model-a")
	require.NoError(t, err)
	assert.Equal(t, []hub.Ref{{Name: "main", Commit: commit1}, {Name: "v1", Commit: commit1}}, refs)
}

func TestListModelsAndAuthor(t *testing.T) {
	var models []*hubtest.Model
	for _, id := range []string{"Org/a", "Org/b", "Org/c", "Org/d", "Org/e", "Other/x"} {
		models = append(models, &hubtest.Model{ID: id})
	}
	h := hubtest.New(models...)
	defer h.Close()
	c := newClient(t, h, "")
	ctx := context.Background()

	ids, err := c.ListModels(ctx, "Org")
	require.NoError(t, err)
	assert.Equal(t, []string{"Org/a", "Org/b", "Org/c", "Org/d", "Org/e"}, ids)
	assert.Equal(t, 3, h.Calls("list"))

	name, err := c.Author(ctx, "org")
	require.NoError(t, err)
	assert.Equal(t, "Org", name)

	_, err = c.Author(ctx, "nobody")
	assert.True(t, errors.IsNotFoundErr(err), err)
}

func TestListModelsForeignNextLink(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Link", `<https://evil.example/api/models?cursor=1>; rel="next"`)
		_, _ = w.Write([]byte(`[{"id":"a/b"}]`))
	}))
	defer s.Close()
	c, err := hub.New(s.URL, hub.Options{Token: "secret"})
	require.NoError(t, err)
	_, err = c.ListModels(context.Background(), "a")
	assert.ErrorContains(t, err, "leaves the hub endpoint")
}

func TestOpen(t *testing.T) {
	h := hubtest.New(fakeModel())
	defer h.Close()
	c := newClient(t, h, "")
	ctx := context.Background()
	content := strings.Repeat("w", 1000)

	cases := []struct {
		name           string
		path           string
		offset, length int64
		want           string
	}{
		{"whole", "model.safetensors", 0, -1, content},
		{"range", "model.safetensors", 10, 20, content[10:30]},
		{"to end", "model.safetensors", 990, -1, content[990:]},
		{"exact size", "model.safetensors", 0, 1000, content},
		{"empty", "model.safetensors", 5, 0, ""},
		{"escaped path", "dir/a b.txt", 0, -1, "spaces"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := c.Open(ctx, "Org/Model-A", commit1, tc.path, tc.offset, tc.length)
			require.NoError(t, err)
			defer r.Close()
			got, err := io.ReadAll(r)
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(got))
		})
	}

	_, err := c.Open(ctx, "Org/Model-A", commit1, "missing", 0, -1)
	assert.True(t, errors.IsNotFoundErr(err), err)
	_, err = c.Open(ctx, "Org/Model-A", "main", "config.json", 0, -1)
	assert.Error(t, err)
}

func TestContentDigest(t *testing.T) {
	m := fakeModel()
	h := hubtest.New(m)
	defer h.Close()
	c := newClient(t, h, "")
	ctx := context.Background()

	s, err := c.Snapshot(ctx, m.ID, commit1)
	require.NoError(t, err)
	var cfg hub.File
	for _, f := range s.Files {
		if f.Path == "config.json" {
			cfg = f
		}
	}
	d, err := c.ContentDigest(ctx, m.ID, commit1, cfg)
	require.NoError(t, err)
	sum := sha256.Sum256([]byte(`{"a":1}`))
	assert.Equal(t, "sha256:"+hex.EncodeToString(sum[:]), d.String())

	bad := cfg
	bad.BlobID = hub.GitBlobID([]byte("other"))
	_, err = c.ContentDigest(ctx, m.ID, commit1, bad)
	assert.ErrorContains(t, err, "git blob id")

	short := cfg
	short.Size = cfg.Size + 1
	_, err = c.ContentDigest(ctx, m.ID, commit1, short)
	assert.Error(t, err)

	huge := cfg
	huge.Size = hub.MaxNonLFSSize() + 1
	calls := h.Calls("resolve")
	_, err = c.ContentDigest(ctx, m.ID, commit1, huge)
	assert.ErrorContains(t, err, "config.json")
	assert.Equal(t, calls, h.Calls("resolve"))

	lfs := hub.File{Path: "x", SHA256: strings.Repeat("a", 64)}
	d, err = c.ContentDigest(ctx, m.ID, commit1, lfs)
	require.NoError(t, err)
	assert.Equal(t, "sha256:"+strings.Repeat("a", 64), d.String())
}

func TestGitBlobID(t *testing.T) {
	// git hash-object of an empty file and of "hello\n"
	assert.Equal(t, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", hub.GitBlobID(nil))
	assert.Equal(t, "ce013625030ba8dba906f756967f9e9ca394464a", hub.GitBlobID([]byte("hello\n")))
}

func TestErrorCodes(t *testing.T) {
	cases := []struct {
		status int
		code   string
	}{
		{http.StatusUnauthorized, errors.UnAuthorizedCode}, // carries X-Error-Code
		{http.StatusForbidden, errors.ForbiddenCode},
		{http.StatusNotFound, errors.NotFoundCode},
		{http.StatusTooManyRequests, errors.RateLimitCode},
		{http.StatusInternalServerError, errors.GeneralCode},
	}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.Header().Set("Retry-After", "0")
				w.Header().Set("X-Error-Code", "SomeCode")
				w.WriteHeader(tc.status)
			}))
			defer s.Close()
			c, err := hub.New(s.URL, hub.Options{})
			require.NoError(t, err)
			_, err = c.Refs(context.Background(), "a/b")
			assert.Equal(t, tc.code, errors.ErrCode(err))
			assert.Contains(t, err.Error(), "SomeCode")
			if tc.status == http.StatusTooManyRequests {
				assert.Equal(t, 2, calls, "429 is retried once")
			} else {
				assert.Equal(t, 1, calls)
			}
		})
	}
}

func TestUnauthorizedMapping(t *testing.T) {
	cases := []struct {
		name      string
		token     string
		errorCode string
		want      string
	}{
		{"anonymous bare 401 is not found", "", "", errors.NotFoundCode},
		{"anonymous gated repo", "", "GatedRepo", errors.UnAuthorizedCode},
		{"401 with a token", "hf_x", "", errors.UnAuthorizedCode},
		{"gated with a token", "hf_x", "GatedRepo", errors.UnAuthorizedCode},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.errorCode != "" {
					w.Header().Set("X-Error-Code", tc.errorCode)
				}
				w.Header().Set("X-Error-Message", "Invalid username or password.")
				w.WriteHeader(http.StatusUnauthorized)
			}))
			defer s.Close()
			c, err := hub.New(s.URL, hub.Options{Token: tc.token})
			require.NoError(t, err)
			_, err = c.Refs(context.Background(), "a/b")
			assert.Equal(t, tc.want, errors.ErrCode(err), err)
		})
	}
}

func TestOpenHeaderTimeout(t *testing.T) {
	release := make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()
	defer close(release)
	defer hub.SetDownloadHeaderTimeout(50 * time.Millisecond)()
	c, err := hub.New(s.URL, hub.Options{})
	require.NoError(t, err)
	_, err = c.Open(context.Background(), "a/b", commit1, "x", 0, -1)
	assert.ErrorContains(t, err, "timed out after 50ms waiting for response headers")
}

func TestRateLimitRetry(t *testing.T) {
	h := hubtest.New(fakeModel())
	defer h.Close()
	c := newClient(t, h, "")

	h.RateLimit("refs", 1)
	_, err := c.Refs(context.Background(), "Org/Model-A")
	require.NoError(t, err)
	assert.Equal(t, 2, h.Calls("refs"))

	h.RateLimit("refs", 2)
	_, err = c.Refs(context.Background(), "Org/Model-A")
	assert.True(t, errors.IsRateLimitError(err), err)
}

func TestRetryAfterTooLong(t *testing.T) {
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer s.Close()
	c, err := hub.New(s.URL, hub.Options{})
	require.NoError(t, err)
	_, err = c.Refs(context.Background(), "a/b")
	assert.True(t, errors.IsRateLimitError(err), err)
	assert.Equal(t, 1, calls)
}

func TestTokenAndGated(t *testing.T) {
	m := fakeModel()
	m.Gated = true
	h := hubtest.New(m)
	h.Token = "hf_secret"
	defer h.Close()
	ctx := context.Background()

	anon := newClient(t, h, "")
	assert.False(t, anon.HasToken())
	// A bare anonymous 401 maps to not found; HealthCheck never calls WhoAmI without a token.
	assert.Equal(t, errors.NotFoundCode, errors.ErrCode(anon.WhoAmI(ctx)))
	_, err := anon.Refs(ctx, m.ID)
	assert.Equal(t, errors.UnAuthorizedCode, errors.ErrCode(err))
	assert.NoError(t, anon.Ping(ctx))

	authed := newClient(t, h, "hf_secret")
	assert.True(t, authed.HasToken())
	assert.NoError(t, authed.WhoAmI(ctx))
	_, err = authed.Refs(ctx, m.ID)
	assert.NoError(t, err)
	r, err := authed.Open(ctx, m.ID, commit1, "config.json", 0, -1)
	require.NoError(t, err)
	r.Close()
}

func TestNewInvalidEndpoint(t *testing.T) {
	for _, e := range []string{"", "huggingface.co", "://x"} {
		_, err := hub.New(e, hub.Options{})
		assert.Error(t, err, e)
	}
}

func TestIsCommit(t *testing.T) {
	assert.True(t, hub.IsCommit(commit1))
	assert.False(t, hub.IsCommit("main"))
	assert.False(t, hub.IsCommit(strings.ToUpper("abcdefabcdefabcdefabcdefabcdefabcdefabcd")))
	assert.False(t, hub.IsCommit(commit1+"1"))
}
