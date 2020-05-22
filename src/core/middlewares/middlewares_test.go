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

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_fetchBlobAPISkipper(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"fetch blob", args{httptest.NewRequest(http.MethodGet, "/v2/library/photon/blobs/sha256:6e0447537050cf871f9ab6a3fec5715f9c6fff5212f6666993f1fc46b1f717a3", nil)}, true},
		{"delete blob", args{httptest.NewRequest(http.MethodDelete, "/v2/library/photon/blobs/sha256:6e0447537050cf871f9ab6a3fec5715f9c6fff5212f6666993f1fc46b1f717a3", nil)}, false},
		{"get manifest", args{httptest.NewRequest(http.MethodDelete, "/v2/library/photon/manifests/sha256:6e0447537050cf871f9ab6a3fec5715f9c6fff5212f6666993f1fc46b1f717a3", nil)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fetchBlobAPISkipper(tt.args.r); got != tt.want {
				t.Errorf("fetchBlobAPISkipper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_legacyAPISkipper(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"/api/v2.0/projects", args{httptest.NewRequest(http.MethodGet, "/api/v2.0/projects", nil)}, false},
		{"//api/v2.0/projects", args{httptest.NewRequest(http.MethodGet, "//api/v2.0/projects", nil)}, false},
		{"/api/v2.0//projects", args{httptest.NewRequest(http.MethodGet, "/api/v2.0//projects", nil)}, false},
		{"/v2/library/photon/tags", args{httptest.NewRequest(http.MethodGet, "/v2/library/photon/tags", nil)}, false},
		{"/api/projects", args{httptest.NewRequest(http.MethodGet, "/api/projects", nil)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := legacyAPISkipper(tt.args.r); got != tt.want {
				t.Errorf("legacyAPISkipper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readonlySkipper(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"login", args{httptest.NewRequest(http.MethodPost, "/c/login", nil)}, true},
		{"login get", args{httptest.NewRequest(http.MethodGet, "/c/login", nil)}, false},
		{"onboard", args{httptest.NewRequest(http.MethodPost, "/c/oidc/onboard", nil)}, true},
		{"user exist", args{httptest.NewRequest(http.MethodPost, "/c/userExists", nil)}, true},
		{"user exist", args{httptest.NewRequest(http.MethodPost, "/service/notifications/jobs/adminjob/123456", nil)}, true},
		{"user exist", args{httptest.NewRequest(http.MethodPost, "/service/notifications/jobs/adminjob/abcdefg", nil)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pass bool
			for _, skipper := range readonlySkippers {
				if got := skipper(tt.args.r); got == tt.want {
					pass = true
				}
			}
			if !pass {
				t.Errorf("readonlySkippers() = %v, want %v", tt.args, tt.want)
			}
		})
	}
}
