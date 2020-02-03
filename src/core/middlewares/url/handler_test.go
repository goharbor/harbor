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
// limitations under the License

package url

import (
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if result := m.Run(); result != 0 {
		os.Exit(result)
	}
}

func TestParseURL(t *testing.T) {
	cases := []struct {
		input  string
		expect map[string]string
		match  bool
	}{
		{
			input:  "/api/projects",
			expect: map[string]string{},
			match:  false,
		},
		{
			input:  "/v2/_catalog",
			expect: map[string]string{},
			match:  false,
		},
		{
			input: "/v2/no-project-repo/tags/list",
			expect: map[string]string{
				util.RepositorySubexp: "no-project-repo",
			},
			match: true,
		},
		{
			input: "/v2/development/golang/manifests/sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			expect: map[string]string{
				util.RepositorySubexp: "development/golang",
				util.ReferenceSubexp:  "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				util.DigestSubexp:     "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			},
			match: true,
		},
		{
			input:  "/v2/development/golang/manifests/shaxxx:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			expect: map[string]string{},
			match:  false,
		},
		{
			input: "/v2/multi/sector/repository/blobs/sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			expect: map[string]string{
				util.RepositorySubexp: "multi/sector/repository",
				util.DigestSubexp:     "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			},
			match: true,
		},
		{
			input:  "/v2/blobs/uploads",
			expect: map[string]string{},
			match:  false,
		},
		{
			input: "/v2/library/ubuntu/blobs/uploads",
			expect: map[string]string{
				util.RepositorySubexp: "library/ubuntu",
			},
			match: true,
		},
		{
			input: "/v2/library/centos/blobs/uploads/u-12345",
			expect: map[string]string{
				util.RepositorySubexp: "library/centos",
			},
			match: true,
		},
	}

	for _, c := range cases {
		e, m := parse(c.input)
		assert.Equal(t, c.match, m)
		assert.Equal(t, c.expect, e)
	}
}
