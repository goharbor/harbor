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

package event

import (
	"context"
	"testing"
)

func TestEventResolver_PreCheck(t *testing.T) {
	type fields struct {
		ResourceType        string
		SucceedCodes        []int
		SensitiveAttributes []string
		ShouldResolveName   bool
		IDToNameFunc        ResolveIDToNameFunc
		ResourceIDPattern   string
	}
	type args struct {
		ctx    context.Context
		url    string
		method string
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantCapture      bool
		wantResourceName string
	}{
		{"test normal", fields{ResourceIDPattern: `/api/v2.0/tests/(\d+)`, ResourceType: "test", SucceedCodes: []int{200}, ShouldResolveName: true, IDToNameFunc: func(string) string { return "test" }}, args{context.Background(), "/api/v2.0/tests/123", "DELETE"}, true, "test"},
		{"test resource name", fields{ResourceIDPattern: `/api/v2.0/tests/(\d+)`, ResourceType: "test", SucceedCodes: []int{200}, ShouldResolveName: true, IDToNameFunc: func(string) string { return "test_resource_name" }}, args{context.Background(), "/api/v2.0/tests/234", "DELETE"}, true, "test_resource_name"},
		{"test no resource name", fields{ResourceIDPattern: `/api/v2.0/tests/(\d+)`, ResourceType: "test", SucceedCodes: []int{200}, ShouldResolveName: true}, args{context.Background(), "/api/v2.0/tests/234", "GET"}, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Resolver{
				ResourceIDPattern:   tt.fields.ResourceIDPattern,
				ResourceType:        tt.fields.ResourceType,
				SucceedCodes:        tt.fields.SucceedCodes,
				SensitiveAttributes: tt.fields.SensitiveAttributes,
				ShouldResolveName:   tt.fields.ShouldResolveName,
				IDToNameFunc:        tt.fields.IDToNameFunc,
			}
			gotCapture, gotResourceName := e.PreCheck(tt.args.ctx, tt.args.url, tt.args.method)
			if gotCapture != tt.wantCapture {
				t.Errorf("EventResolver.PreCheck() gotCapture = %v, want %v", gotCapture, tt.wantCapture)
			}
			if gotResourceName != tt.wantResourceName {
				t.Errorf("EventResolver.PreCheck() gotResourceName = %v, want %v", gotResourceName, tt.wantResourceName)
			}
		})
	}
}
