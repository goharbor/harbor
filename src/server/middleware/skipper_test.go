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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"
)

func TestMethodAndPathSkipper(t *testing.T) {
	type args struct {
		method string
		re     *regexp.Regexp
		r      *http.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"match method and path", args{http.MethodGet, regexp.MustCompile(`/req`), httptest.NewRequest(http.MethodGet, "/req", nil)}, true},
		{"match method only", args{http.MethodGet, regexp.MustCompile(`/req`), httptest.NewRequest(http.MethodGet, "/path", nil)}, false},
		{"match path only", args{http.MethodGet, regexp.MustCompile(`/req`), httptest.NewRequest(http.MethodPost, "/req", nil)}, false},
		{"match all methods", args{"*", regexp.MustCompile(`/req`), httptest.NewRequest(http.MethodPost, "/req", nil)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MethodAndPathSkipper(tt.args.method, tt.args.re)(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MethodAndPathSkipper()() = %v, want %v", got, tt.want)
			}
		})
	}
}
