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

package trace

import (
	"net/http"
	"net/url"
	"testing"
)

func TestHarborSpanNameFormatter(t *testing.T) {
	type args struct {
		operation string
		request   *http.Request
	}
	normalReq, err := http.NewRequest("GET", "http://10.192.168.1/api/v2.0/configuration", nil)
	if err != nil {
		t.Error(err)
	}
	normalReqWithHTTPS, err := http.NewRequest("GET", "https://10.192.168.1/api/v2.0/configuration", nil)
	if err != nil {
		t.Error(err)
	}
	cases := []struct {
		name string
		in   args
		want string
	}{
		{
			name: `normal`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodGet, URL: &url.URL{Scheme: "http", Path: "/api/v2.0/configuration"}}},
			want: "GET http://10.192.168.1/api/v2.0/configuration",
		},
		{
			name: `normal request `,
			in:   args{"sample", normalReq},
			want: "GET http://10.192.168.1/api/v2.0/configuration",
		},
		{
			name: `normal request with https `,
			in:   args{"sample", normalReqWithHTTPS},
			want: "GET https://10.192.168.1/api/v2.0/configuration",
		},
		{
			name: `no host`,
			in:   args{"sample", &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Path: "/api/v2.0/configuration"}}},
			want: "GET http://host_unknown/api/v2.0/configuration",
		},
		{
			name: `no schema`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodGet, URL: &url.URL{Path: "/api/v2.0/configuration"}}},
			want: "GET http://10.192.168.1/api/v2.0/configuration",
		},
		{
			name: `https`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodGet, URL: &url.URL{Scheme: "https", Path: "/api/v2.0/configuration"}}},
			want: "GET https://10.192.168.1/api/v2.0/configuration",
		},
		{
			name: `empty path`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodGet, URL: &url.URL{Scheme: "http"}}},
			want: "sample",
		},
		{
			name: `nil url`,
			in:   args{"sample", &http.Request{Host: "10.192.168.1", Method: http.MethodGet}},
			want: "sample",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := HarborSpanNameFormatter(tt.in.operation, tt.in.request)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
