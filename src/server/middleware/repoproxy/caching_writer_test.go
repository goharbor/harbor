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

package repoproxy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachingResponseWriter_CacheMiss(t *testing.T) {
	digest := "sha256:test-cache-miss"
	rec := httptest.NewRecorder()
	cw := NewCachingResponseWriter(rec, digest)

	data := []byte("hello world")
	cw.WriteHeader(http.StatusOK)
	n, err := cw.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Set Content-Length as the upstream would.
	cw.Header().Set("Content-Length", strconv.Itoa(len(data)))
	cw.Close()

	// Verify blob was cached to disk.
	finalPath := filepath.Join(os.TempDir(), "lru_blob_cache", digest)
	stat, err := os.Stat(finalPath)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(data)), stat.Size())

	// Cleanup.
	os.Remove(finalPath)
}

func TestCachingResponseWriter_NonSuccessSkipsCache(t *testing.T) {
	digest := "sha256:test-non-success"
	rec := httptest.NewRecorder()
	cw := NewCachingResponseWriter(rec, digest)

	cw.WriteHeader(http.StatusNotFound)
	cw.Write([]byte("not found"))
	cw.Close()

	// Blob should NOT be cached.
	finalPath := filepath.Join(os.TempDir(), "lru_blob_cache", digest)
	_, err := os.Stat(finalPath)
	assert.True(t, os.IsNotExist(err))
}

func TestCachingResponseWriter_TempFileIsUnique(t *testing.T) {
	digest := "sha256:test-unique-temp"
	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()

	cw1 := NewCachingResponseWriter(rec1, digest)
	cw2 := NewCachingResponseWriter(rec2, digest)

	// Two writers for the same digest should use different temp files.
	assert.NotEqual(t, cw1.tempPath, cw2.tempPath)

	cw1.Close()
	cw2.Close()
}
