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

package sizequota

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_parseUploadedBlobSize(t *testing.T) {
	writer := func(header string) http.ResponseWriter {
		rr := httptest.NewRecorder()
		if header != "" {
			rr.Header().Add("Range", header)
		}
		return rr
	}
	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"success", args{writer("0-99")}, 100, false},
		{"ranage header not found", args{writer("")}, 0, true},
		{"ranage header bad value", args{writer("0")}, 0, true},
		{"ranage header bad value", args{writer("0-")}, 0, true},
		{"ranage header bad value", args{writer("0-a")}, 0, true},
		{"ranage header bad value", args{writer("0-1-2")}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUploadedBlobSize(tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUploadedBlobSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseUploadedBlobSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
