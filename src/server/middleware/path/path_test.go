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

package path

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func Test_escape(t *testing.T) {
	re := regexp.MustCompile(`/api/v2.0/projects/.*/repositories/(.*)/artifacts`)

	type args struct {
		re   *regexp.Regexp
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"/api/v2.0/projects/library/repositories/photon/artifacts",
			args{re, "/api/v2.0/projects/library/repositories/photon/artifacts"},
			"/api/v2.0/projects/library/repositories/photon/artifacts",
		},
		{
			"/api/v2.0/projects/library/repositories/photon/hello-world/artifacts",
			args{re, "/api/v2.0/projects/library/repositories/photon/hello-world/artifacts"},
			"/api/v2.0/projects/library/repositories/photon%2Fhello-world/artifacts",
		},
		{
			"/api/v2.0/projects/library/repositories/photon/hello-world/artifacts/digest/scan",
			args{re, "/api/v2.0/projects/library/repositories/photon/hello-world/artifacts/digest/scan"},
			"/api/v2.0/projects/library/repositories/photon%2Fhello-world/artifacts/digest/scan",
		},

		{
			"/api/v2.0/projects/library/repositories",
			args{re, "/api/v2.0/projects/library/repositories"},
			"/api/v2.0/projects/library/repositories",
		},
		{
			"/api/v2.0/projects/library/repositories/hello/mariadb/_self",
			args{regexp.MustCompile(`^/api/v2.0/projects/.*/repositories/(.*)/_self`), "/api/v2.0/projects/library/repositories/hello/mariadb/_self"},
			"/api/v2.0/projects/library/repositories/hello%2Fmariadb/_self",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escape(tt.args.re, tt.args.path); got != tt.want {
				t.Errorf("escape() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscapeMiddleware(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v2.0/projects/library/repositories/hello/mariadb/_self", nil)
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2.0/projects/library/repositories/hello%2Fmariadb/_self" {
			t.Errorf("escape middleware failed")
		}
		w.WriteHeader(http.StatusOK)
	})

	EscapeMiddleware()(next).ServeHTTP(w, r)
}
