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

package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRebaseOnCore(t *testing.T) {
	const hook = "/service/notifications/tasks/42"

	cases := []struct {
		name    string
		coreURL string
		raw     string
		want    string
	}{
		{
			name:    "http to https after internal TLS enabled",
			coreURL: "https://harbor-core:443",
			raw:     "http://harbor-core:80" + hook,
			want:    "https://harbor-core:443" + hook,
		},
		{
			name:    "https to http after internal TLS disabled",
			coreURL: "http://harbor-core:80",
			raw:     "https://harbor-core:443" + hook,
			want:    "http://harbor-core:80" + hook,
		},
		{
			name:    "unchanged when CORE_URL matches",
			coreURL: "https://harbor-core:443",
			raw:     "https://harbor-core:443" + hook,
			want:    "https://harbor-core:443" + hook,
		},
		{
			name:    "trailing slash on CORE_URL is ignored",
			coreURL: "https://harbor-core:443/",
			raw:     "http://harbor-core:80" + hook,
			want:    "https://harbor-core:443" + hook,
		},
		{
			name:    "unchanged when CORE_URL is empty",
			coreURL: "",
			raw:     "http://harbor-core:80" + hook,
			want:    "http://harbor-core:80" + hook,
		},
		{
			name:    "unchanged when CORE_URL is not a URL",
			coreURL: "://broken",
			raw:     "http://harbor-core:80" + hook,
			want:    "http://harbor-core:80" + hook,
		},
		{
			name:    "unchanged when CORE_URL has no scheme",
			coreURL: "//harbor-core:443",
			raw:     "http://harbor-core:80" + hook,
			want:    "http://harbor-core:80" + hook,
		},
		{
			name:    "unchanged when CORE_URL carries a path prefix",
			coreURL: "https://proxy/core",
			raw:     "http://harbor-core:80" + hook,
			want:    "http://harbor-core:80" + hook,
		},
		{
			name:    "unchanged when the hook URL has no scheme",
			coreURL: "https://harbor-core:443",
			raw:     "//harbor-core:80" + hook,
			want:    "//harbor-core:80" + hook,
		},
		{
			name:    "unchanged when the hook URL is not absolute",
			coreURL: "https://harbor-core:443",
			raw:     hook,
			want:    hook,
		},
		{
			name:    "unchanged when the hook URL is garbage",
			coreURL: "https://harbor-core:443",
			raw:     "not a url",
			want:    "not a url",
		},
		{
			name:    "unchanged when the hook URL is not http or https",
			coreURL: "https://harbor-core:443",
			raw:     "ftp://harbor-core:80" + hook,
			want:    "ftp://harbor-core:80" + hook,
		},
		{
			name:    "unchanged when the path is not a core notification",
			coreURL: "https://harbor-core:443",
			raw:     "http://elsewhere:8080/webhook",
			want:    "http://elsewhere:8080/webhook",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("CORE_URL", c.coreURL)
			assert.Equal(t, c.want, rebaseOnCore(c.raw))
		})
	}
}
