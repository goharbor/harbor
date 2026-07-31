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
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnapshot(t *testing.T) {
	var token string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = r.Header.Get("Authorization")
		switch r.URL.Path {
		case "/api/models/org/model/revision/main":
			_, _ = w.Write([]byte(`{
				"sha":"abc123",
				"siblings":[
					{"rfilename":"README.md","size":64},
					{"rfilename":"weights/model.safetensors","size":128}
				]
			}`))
		case "/org/model/resolve/main/README.md":
			_, _ = w.Write([]byte("---\nlicense: apache-2.0\ntags:\n- text-generation\n---\n# Model\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := &client{
		endpoint:   server.URL,
		httpClient: server.Client(),
	}
	snapshot, err := c.Snapshot(context.Background(), "org/model", "main", "", "secret")
	require.NoError(t, err)
	require.Equal(t, "Bearer secret", token)
	require.Equal(t, "abc123", snapshot.CommitSHA)
	require.Len(t, snapshot.Files, 2)
	require.Equal(t, "apache-2.0", snapshot.CardData["license"])
	require.Contains(t, string(snapshot.Readme), "# Model")
}

func TestSnapshotFiltersSubpath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/models/org/model/revision/main" {
			_, _ = w.Write([]byte(`{
				"sha":"abc123",
				"siblings":[
					{"rfilename":"adapter/README.md","size":64},
					{"rfilename":"adapter/model.bin","size":128},
					{"rfilename":"other/model.bin","size":128}
				]
			}`))
			return
		}
		if r.URL.Path == "/org/model/resolve/main/adapter/README.md" {
			_, _ = w.Write([]byte("# Adapter\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := &client{
		endpoint:   server.URL,
		httpClient: server.Client(),
	}
	snapshot, err := c.Snapshot(context.Background(), "org/model", "main", "adapter", "")
	require.NoError(t, err)
	require.Len(t, snapshot.Files, 2)
	require.Equal(t, "README.md", snapshot.Files[0].Path)
	require.Equal(t, "model.bin", snapshot.Files[1].Path)
}

func TestOpenFileRetriesTransientEOF(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			hj, ok := w.(http.Hijacker)
			require.True(t, ok)
			conn, _, err := hj.Hijack()
			require.NoError(t, err)
			_ = conn.Close()
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	c := &client{
		endpoint:   server.URL,
		fileClient: server.Client(),
	}
	reader, err := c.OpenFile(context.Background(), "org/model", "main", "weights.bin", "")
	require.NoError(t, err)
	defer reader.Close()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "ok", string(data))
	require.GreaterOrEqual(t, attempts.Load(), int32(2))
}

func TestIsTransientNetErr(t *testing.T) {
	require.True(t, isTransientNetErr(io.ErrUnexpectedEOF))
	require.True(t, isTransientNetErr(errors.New(`Get "https://huggingface.co/x": unexpected EOF`)))
	require.False(t, isTransientNetErr(errors.New("download weights.bin failed: 404 Not Found")))
}
