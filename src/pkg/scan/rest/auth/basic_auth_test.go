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
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_basicAuthorizer_Authorize(t *testing.T) {
	type fields struct {
		typeID     string
		accessCred string
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"ok", fields{"Basic", "username:password"}, args{httptest.NewRequest("GET", "/", nil)}, false},
		{"empty cerd", fields{"Basic", ""}, args{httptest.NewRequest("GET", "/", nil)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := &basicAuthorizer{
				typeID:     tt.fields.typeID,
				accessCred: tt.fields.accessCred,
			}
			if err := ba.Authorize(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("basicAuthorizer.Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
