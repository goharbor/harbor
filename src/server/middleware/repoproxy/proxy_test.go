//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	securitySecret "github.com/goharbor/harbor/src/common/security/secret"
)

func TestIsProxySession(t *testing.T) {
	sc1 := securitySecret.NewSecurityContext("123456789", nil)
	otherCtx := security.NewContext(context.Background(), sc1)

	sc2 := proxycachesecret.NewSecurityContext("library/hello-world")
	proxyCtx := security.NewContext(context.Background(), sc2)

	user := &models.User{
		Username: "robot$library+scanner-8ec3b47a-fd29-11ee-9681-0242c0a87009",
	}
	userSc := local.NewSecurityContext(user)
	scannerCtx := security.NewContext(context.Background(), userSc)

	otherRobot := &models.User{
		Username: "robot$library+test-8ec3b47a-fd29-11ee-9681-0242c0a87009",
	}
	userSc2 := local.NewSecurityContext(otherRobot)
	nonScannerCtx := security.NewContext(context.Background(), userSc2)

	cases := []struct {
		name string
		in   context.Context
		want bool
	}{
		{
			name: `normal`,
			in:   otherCtx,
			want: false,
		},
		{
			name: `proxy user`,
			in:   proxyCtx,
			want: true,
		},
		{
			name: `robot account`,
			in:   scannerCtx,
			want: true,
		},
		{
			name: `non scanner robot`,
			in:   nonScannerCtx,
			want: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := isProxySession(tt.in, "library")
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}

func TestMatchRepositoryFilter(t *testing.T) {
	cases := []struct {
		name          string
		repository    string
		filterPattern string
		filterKind    string
		want          bool
	}{
		{
			name:          "empty pattern matches all",
			repository:    "library/nginx",
			filterPattern: "",
			filterKind:    "",
			want:          true,
		},
		{
			name:          "regex exact match",
			repository:    "library/nginx",
			filterPattern: "^library/nginx$",
			filterKind:    "regex",
			want:          true,
		},
		{
			name:          "regex partial match is anchored",
			repository:    "library/nginx",
			filterPattern: "nginx",
			filterKind:    "regex",
			want:          false,
		},
		{
			name:          "regex prefix match is anchored",
			repository:    "library/nginx",
			filterPattern: "^library/",
			filterKind:    "regex",
			want:          false,
		},
		{
			name:          "regex wildcard match",
			repository:    "library/nginx",
			filterPattern: "^library/.*",
			filterKind:    "regex",
			want:          true,
		},
		{
			name:          "regex no match",
			repository:    "library/nginx",
			filterPattern: "^other/",
			filterKind:    "regex",
			want:          false,
		},
		{
			name:          "regex invalid pattern treated as no match",
			repository:    "library/nginx",
			filterPattern: "[invalid",
			filterKind:    "regex",
			want:          false,
		},
		{
			name:          "doublestar exact match",
			repository:    "library/nginx",
			filterPattern: "library/nginx",
			filterKind:    "doublestar",
			want:          true,
		},
		{
			name:          "doublestar single star",
			repository:    "library/nginx",
			filterPattern: "library/*",
			filterKind:    "doublestar",
			want:          true,
		},
		{
			name:          "doublestar double star",
			repository:    "org/team/repo",
			filterPattern: "org/**",
			filterKind:    "doublestar",
			want:          true,
		},
		{
			name:          "doublestar match all",
			repository:    "library/nginx",
			filterPattern: "**",
			filterKind:    "doublestar",
			want:          true,
		},
		{
			name:          "doublestar no match",
			repository:    "library/nginx",
			filterPattern: "other/*",
			filterKind:    "doublestar",
			want:          false,
		},
		{
			name:          "doublestar alternation",
			repository:    "library/nginx",
			filterPattern: "library/{nginx,alpine}",
			filterKind:    "doublestar",
			want:          true,
		},
		{
			name:          "empty kind defaults to doublestar",
			repository:    "library/nginx",
			filterPattern: "library/**",
			filterKind:    "",
			want:          true,
		},
		{
			name:          "empty pattern with kind set matches all",
			repository:    "library/nginx",
			filterPattern: "",
			filterKind:    "regex",
			want:          true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := matchRepositoryFilter(tt.repository, tt.filterPattern, tt.filterKind)
			if got != tt.want {
				t.Errorf("matchRepositoryFilter(%q, %q, %q) = %v; want %v",
					tt.repository, tt.filterPattern, tt.filterKind, got, tt.want)
			}
		})
	}
}
