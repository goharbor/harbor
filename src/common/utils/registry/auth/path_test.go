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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRepository(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"/v2", ""},
		{"/v2/_catalog", ""},
		{"/v2/library/tags/list", "library"},
		{"/v2/tags/list", ""},
		{"/v2/tags/list/tags/list", "tags/list"},
		{"/v2/library/manifests/latest", "library"},
		{"/v2/library/manifests/sha256:eec76eedea59f7bf39a2713bfd995c82cfaa97724ee5b7f5aba253e07423d0ae", "library"},
		{"/v2/library/blobs/sha256:eec76eedea59f7bf39a2713bfd995c82cfaa97724ee5b7f5aba253e07423d0ae", "library"},
		{"/v2/library/blobs/uploads", "library"},
		{"/v2/library/blobs/uploads/1234567890", "library"},
	}

	for _, c := range cases {
		assert.Equal(t, c.output, parseRepository(c.input))
	}
}
