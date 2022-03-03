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

package security

import (
	"net/http"
	"testing"

	_ "github.com/goharbor/harbor/src/core/auth/db"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuth(t *testing.T) {
	basicAuth := &basicAuth{}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req.SetBasicAuth("admin", "Harbor12345")
	req = req.WithContext(orm.Context())
	ctx := basicAuth.Generate(req)
	assert.NotNil(t, ctx)
}

func TestGetClientIP(t *testing.T) {
	h := http.Header{}
	h.Set("X-Forwarded-For", "1.1.1.1")
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nil request", args{nil}, ""},
		{"no header", args{&http.Request{RemoteAddr: "10.10.10.10"}}, "10.10.10.10"},
		{"set x forworded for", args{&http.Request{Header: h, RemoteAddr: "10.10.10.10"}}, "1.1.1.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetClientIP(tt.args.r); got != tt.want {
				t.Errorf("GetClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserAgent(t *testing.T) {
	h := http.Header{}
	h.Set("user-agent", "docker")
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nil request", args{nil}, ""},
		{"no header", args{&http.Request{}}, ""},
		{"with user-agent", args{&http.Request{Header: h}}, "docker"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserAgent(tt.args.r); got != tt.want {
				t.Errorf("GetUserAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}
