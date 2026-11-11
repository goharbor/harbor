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

package lib

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchManifestURLPattern(t *testing.T) {
	_, _, ok := MatchManifestURLPattern("/v2/library/hello-world/manifests/.Invalid")
	assert.True(t, ok)

	_, _, ok = MatchManifestURLPattern("/v2/")
	assert.False(t, ok)

	// the reference is a single path segment
	_, _, ok = MatchManifestURLPattern("/v2/library/hello-world/manifests//")
	assert.False(t, ok)

	_, _, ok = MatchManifestURLPattern("/v2/library/hello-world/manifests/###")
	assert.True(t, ok)

	repository, reference, ok := MatchManifestURLPattern("/v2/library/hello-world/manifests/latest")
	assert.True(t, ok)
	assert.Equal(t, "library/hello-world", repository)
	assert.Equal(t, "latest", reference)

	repository, reference, ok = MatchManifestURLPattern("/v2/library/hello-world/manifests/sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9")
	assert.True(t, ok)
	assert.Equal(t, "library/hello-world", repository)
	assert.Equal(t, "sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9", reference)
}

func TestMatchBlobURLPattern(t *testing.T) {
	_, _, ok := MatchBlobURLPattern("")
	assert.False(t, ok)

	_, _, ok = MatchBlobURLPattern("/v2/")
	assert.False(t, ok)

	repository, digest, ok := MatchBlobURLPattern("/v2/library/hello-world/blobs/sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9")
	assert.True(t, ok)
	assert.Equal(t, "library/hello-world", repository)
	assert.Equal(t, "sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9", digest)
}

func TestMatchBlobUploadURLPattern(t *testing.T) {
	_, ok := MatchBlobUploadURLPattern("")
	assert.False(t, ok)

	_, ok = MatchBlobUploadURLPattern("/v2/")
	assert.False(t, ok)

	repository, ok := MatchBlobUploadURLPattern("/v2/library/hello-world/blobs/uploads/")
	assert.True(t, ok)
	assert.Equal(t, "library/hello-world", repository)

	repository, ok = MatchBlobUploadURLPattern("/v2/library/hello-world/blobs/uploads/uuid")
	assert.True(t, ok)
	assert.Equal(t, "library/hello-world", repository)
}

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
			url:   "/v2/_catalog////",
			match: true,
		},
		{
			url:   "/v2/_catalog/xxx",
			match: true,
		},
		{
			url:   "/v2/_catalog////#",
			match: true,
		},
		{
			url:   "/v2/_catalog//#//",
			match: true,
		},
	}
	for _, c := range cases {

		assert.Equal(t, c.match, V2CatalogURLRe.MatchString(c.url), "failed for %s", c.url)
	}
}

func TestMatchReferrersURLPattern(t *testing.T) {
	cases := []struct {
		url   string
		match bool
	}{
		{
			url:   "/v2/library/hello-world/referrers/!@#!@#%",
			match: true,
		},
		{
			url:   "/v2/library/hello-world/referrers/test",
			match: true,
		},
		{
			url:   "/v2/library/hello-world/referrers/sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9",
			match: true,
		},
		{
			url:   "/v2/library/hello-world/referrers/e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9",
			match: true,
		},
		{
			url:   "/v2/library/hello-world/referrers/.Invalid",
			match: true,
		},
		{
			// the reference is a single path segment
			url:   "/v2/library/hello-world/referrers//v2/library/photon/referrers/sha256:0000000000000000000000000000000000000000000000000000000000000000",
			match: false,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.match, V2ReferrersURLRe.MatchString(c.url), "failed for %s", c.url)
	}
}

// TestV2URLPatternsRepositoryWithKeywordComponents checks that a URL whose repository name has the
// components "manifests", "referrers", "tags/list" or "blobs" matches only the pattern of its own API
func TestV2URLPatternsRepositoryWithKeywordComponents(t *testing.T) {
	dgt := "sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9"
	patterns := map[string]*regexp.Regexp{
		"manifest":  V2ManifestURLRe,
		"tag_list":  V2TagListURLRe,
		"referrers": V2ReferrersURLRe,
		"blob":      V2BlobURLRe,
	}
	cases := []struct {
		url        string
		pattern    string
		repository string
	}{
		{"/v2/library/x/manifests/y/blobs/" + dgt, "blob", "library/x/manifests/y"},
		{"/v2/library/x/manifests/blobs/" + dgt, "blob", "library/x/manifests"},
		{"/v2/library/x/referrers/y/blobs/" + dgt, "blob", "library/x/referrers/y"},
		{"/v2/library/x/tags/list/blobs/" + dgt, "blob", "library/x/tags/list"},
		{"/v2/library/x/manifests/y/tags/list", "tag_list", "library/x/manifests/y"},
		{"/v2/library/x/blobs/uploads/tags/list", "tag_list", "library/x/blobs/uploads"},
		{"/v2/library/x/tags/list/manifests/latest", "manifest", "library/x/tags/list"},
		{"/v2/library/x/referrers/manifests/" + dgt, "manifest", "library/x/referrers"},
		{"/v2/library/x/manifests/y/referrers/" + dgt, "referrers", "library/x/manifests/y"},
	}
	for _, c := range cases {
		for name, re := range patterns {
			l := re.FindStringSubmatch(c.url)
			if name != c.pattern {
				assert.Empty(t, l, "%s must not match %s", name, c.url)
				continue
			}
			if assert.NotEmpty(t, l, "%s must match %s", name, c.url) {
				assert.Equal(t, c.repository, l[re.SubexpIndex(RepositorySubexp)], c.url)
			}
		}
	}
}

func TestRepositoryNamePattern(t *testing.T) {
	assert := assert.New(t)
	assert.False(RepositoryNameRe.MatchString("a/*"))
	assert.False(RepositoryNameRe.MatchString("a/"))
	assert.True(RepositoryNameRe.MatchString("a/b"))
	assert.True(RepositoryNameRe.MatchString("a"))
}
