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

package auth

import (
	"regexp"

	"github.com/docker/distribution/reference"
	"github.com/goharbor/harbor/src/common/utils/log"
)

var (
	base            = regexp.MustCompile("/v2")
	catalog         = regexp.MustCompile("/v2/_catalog")
	tag             = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/tags/list")
	manifest        = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/manifests/(" + reference.TagRegexp.String() + "|" + reference.DigestRegexp.String() + ")")
	blob            = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/blobs/" + reference.DigestRegexp.String())
	blobUpload      = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/blobs/uploads")
	blobUploadChunk = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/blobs/uploads/[a-zA-Z0-9-_.=]+")

	repoRegExps = []*regexp.Regexp{tag, manifest, blob, blobUploadChunk, blobUpload}
)

// parse the repository name from path, if the path doesn't match any
// regular expressions in repoRegExps, nil string will be returned
func parseRepository(path string) string {
	for _, regExp := range repoRegExps {
		subs := regExp.FindStringSubmatch(path)
		// no match
		if subs == nil {
			continue
		}

		// match
		// the subs should contain at least 2 matching texts, the first one matches
		// the whole regular expression, and the second one matches the repository
		// part
		if len(subs) < 2 {
			log.Warningf("unexpected length of sub matches: %d, should >= 2 ", len(subs))
			continue
		}
		return subs[1]
	}

	return ""
}
