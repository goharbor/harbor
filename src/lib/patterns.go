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
	"fmt"
	"regexp"

	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
)

const (
	// RepositorySubexp is the name for sub regex that maps to repository name in the url
	RepositorySubexp = "repository"
	// ReferenceSubexp is the name for sub regex that maps to reference (tag or digest) url
	ReferenceSubexp = "reference"
	// DigestSubexp is the name for sub regex that maps to digest in the url
	DigestSubexp = "digest"
)

var (
	// V2ManifestURLRe is the regular expression for matching request v2 handler to view/delete manifest
	V2ManifestURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/manifests/(?P<%s>.*)$`, RepositorySubexp, reference.NameRegexp.String(), ReferenceSubexp))
	// V2TagListURLRe is the regular expression for matching request to v2 handler to list tags
	V2TagListURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/tags/list`, RepositorySubexp, reference.NameRegexp.String()))
	// V2BlobURLRe is the regular expression for matching request to v2 handler to retrieve head/delete a blob
	V2BlobURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/blobs/(?P<%s>%s)$`, RepositorySubexp, reference.NameRegexp.String(), DigestSubexp, digest.DigestRegexp.String()))
	// V2BlobUploadURLRe is the regular expression for matching the request to v2 handler to upload a blob, the upload uuid currently is not put into a group
	V2BlobUploadURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/blobs/uploads[/a-zA-Z0-9\-_\.=]*$`, RepositorySubexp, reference.NameRegexp.String()))
	// V2CatalogURLRe is the regular expression for matching the request to v2 handler to list catalog
	V2CatalogURLRe = regexp.MustCompile(`^/v2/_catalog(/.*)?$`)
	// V2ReferrersURLRe is the regular expression for matching request to v2 handler to list referrers
	V2ReferrersURLRe = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/referrers/(?P<%s>.*)$`, RepositorySubexp, reference.NameRegexp.String(), ReferenceSubexp))
	// RepositoryNameRe is the regular expression for  matching repository name
	RepositoryNameRe = regexp.MustCompile(fmt.Sprintf("^%s$", reference.NameRegexp))
)

// MatchManifestURLPattern checks whether the provided path matches the manifest URL pattern,
// if does, returns the repository and reference as well
func MatchManifestURLPattern(path string) (repository, reference string, match bool) {
	strs := V2ManifestURLRe.FindStringSubmatch(path)
	if len(strs) < 3 {
		return "", "", false
	}
	return strs[1], strs[2], true
}

// MatchBlobURLPattern checks whether the provided path matches the blob URL pattern,
// if does, returns the repository and reference as well
func MatchBlobURLPattern(path string) (repository, digest string, match bool) {
	strs := V2BlobURLRe.FindStringSubmatch(path)
	if len(strs) < 3 {
		return "", "", false
	}
	return strs[1], strs[2], true
}

// MatchBlobUploadURLPattern checks whether the provided path matches the blob upload URL pattern,
// if does, returns the repository as well
func MatchBlobUploadURLPattern(path string) (repository string, match bool) {
	strs := V2BlobUploadURLRe.FindStringSubmatch(path)
	if len(strs) < 2 {
		return "", false
	}
	return strs[1], true
}
