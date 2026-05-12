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

package member

import (
	"context"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

func stubLookups() func() {
	origLookup := lookupMemberFn
	origProject := resolveProjectFn
	lookupMemberFn = func(_, memberID string) (string, string) {
		switch memberID {
		case "10":
			return "developers", "g"
		case "20":
			return "testuser", "u"
		default:
			return memberID, ""
		}
	}
	resolveProjectFn = func(nameOrID string) (int64, string) {
		if nameOrID == "1" {
			return 1, "myproject"
		}
		return 0, nameOrID
	}
	return func() {
		lookupMemberFn = origLookup
		resolveProjectFn = origProject
	}
}

func TestResolver_PreCheck(t *testing.T) {
	cleanup := stubLookups()
	defer cleanup()

	type args struct {
		ctx    context.Context
		url    string
		method string
	}
	tests := []struct {
		name             string
		args             args
		wantCapture      bool
		wantResourceName string
	}{
		{"create", args{context.Background(), "/api/v2.0/projects/1/members", http.MethodPost}, true, ""},
		{"delete group", args{context.Background(), "/api/v2.0/projects/1/members/10", http.MethodDelete}, true, "g:developers"},
		{"delete user", args{context.Background(), "/api/v2.0/projects/1/members/20", http.MethodDelete}, true, "u:testuser"},
		{"update", args{context.Background(), "/api/v2.0/projects/1/members/10", http.MethodPut}, true, ""},
		{"get ignored", args{context.Background(), "/api/v2.0/projects/1/members/10", http.MethodGet}, false, ""},
		{"list ignored", args{context.Background(), "/api/v2.0/projects/1/members", http.MethodGet}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolver{}
			gotCapture, gotName := r.PreCheck(tt.args.ctx, tt.args.url, tt.args.method)
			if gotCapture != tt.wantCapture {
				t.Errorf("PreCheck() capture = %v, want %v", gotCapture, tt.wantCapture)
			}
			if gotName != tt.wantResourceName {
				t.Errorf("PreCheck() resourceName = %q, want %q", gotName, tt.wantResourceName)
			}
		})
	}
}

func TestResolver_Resolve(t *testing.T) {
	cleanup := stubLookups()
	defer cleanup()

	tests := []struct {
		name           string
		metadata       *commonevent.Metadata
		wantNil        bool
		wantOperation  string
		wantResource   string
		wantDesc       string
		wantSuccessful bool
		wantProjectID  int64
	}{
		{
			name: "create group member",
			metadata: &commonevent.Metadata{
				Username:         "admin",
				RequestURL:       "/api/v2.0/projects/1/members",
				RequestMethod:    http.MethodPost,
				ResponseCode:     http.StatusCreated,
				ResponseLocation: "/api/v2.0/projects/1/members/10",
			},
			wantOperation:  "create",
			wantResource:   "developers",
			wantDesc:       "create group member developers in project myproject",
			wantSuccessful: true,
			wantProjectID:  1,
		},
		{
			name: "create user member",
			metadata: &commonevent.Metadata{
				Username:         "admin",
				RequestURL:       "/api/v2.0/projects/1/members",
				RequestMethod:    http.MethodPost,
				ResponseCode:     http.StatusCreated,
				ResponseLocation: "/api/v2.0/projects/1/members/20",
			},
			wantOperation:  "create",
			wantResource:   "testuser",
			wantDesc:       "create user member testuser in project myproject",
			wantSuccessful: true,
			wantProjectID:  1,
		},
		{
			name: "create failed",
			metadata: &commonevent.Metadata{
				Username:         "admin",
				RequestURL:       "/api/v2.0/projects/1/members",
				RequestMethod:    http.MethodPost,
				ResponseCode:     http.StatusForbidden,
				ResponseLocation: "",
			},
			wantOperation:  "create",
			wantResource:   "",
			wantSuccessful: false,
			wantProjectID:  1,
		},
		{
			name: "delete group member",
			metadata: &commonevent.Metadata{
				Username:      "admin",
				RequestURL:    "/api/v2.0/projects/1/members/10",
				RequestMethod: http.MethodDelete,
				ResponseCode:  http.StatusOK,
				ResourceName:  "g:developers",
			},
			wantOperation:  "delete",
			wantResource:   "developers",
			wantDesc:       "delete group member developers from project myproject",
			wantSuccessful: true,
			wantProjectID:  1,
		},
		{
			name: "delete failed",
			metadata: &commonevent.Metadata{
				Username:      "admin",
				RequestURL:    "/api/v2.0/projects/1/members/10",
				RequestMethod: http.MethodDelete,
				ResponseCode:  http.StatusForbidden,
				ResourceName:  "u:testuser",
			},
			wantOperation:  "delete",
			wantResource:   "testuser",
			wantSuccessful: false,
			wantProjectID:  1,
		},
		{
			name: "update member",
			metadata: &commonevent.Metadata{
				Username:      "admin",
				RequestURL:    "/api/v2.0/projects/1/members/10",
				RequestMethod: http.MethodPut,
				ResponseCode:  http.StatusOK,
			},
			wantOperation:  "update",
			wantResource:   "developers",
			wantDesc:       "update group member developers in project myproject",
			wantSuccessful: true,
			wantProjectID:  1,
		},
		{
			name: "update unknown type member",
			metadata: &commonevent.Metadata{
				Username:      "admin",
				RequestURL:    "/api/v2.0/projects/1/members/999",
				RequestMethod: http.MethodPut,
				ResponseCode:  http.StatusOK,
			},
			wantOperation:  "update",
			wantResource:   "999",
			wantDesc:       "update member 999 in project myproject",
			wantSuccessful: true,
			wantProjectID:  1,
		},
		{
			name: "GET ignored",
			metadata: &commonevent.Metadata{
				RequestURL:    "/api/v2.0/projects/1/members",
				RequestMethod: http.MethodGet,
			},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolver{}
			evt := &event.Event{}
			if err := r.Resolve(tt.metadata, evt); err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if tt.wantNil {
				if evt.Data != nil {
					t.Error("Resolve() should not set event data")
				}
				return
			}
			data, ok := evt.Data.(*model.CommonEvent)
			if !ok || data == nil {
				t.Fatal("Resolve() did not set event data")
			}
			if data.Operation != tt.wantOperation {
				t.Errorf("Operation = %v, want %v", data.Operation, tt.wantOperation)
			}
			if data.ResourceType != "member" {
				t.Errorf("ResourceType = %v, want member", data.ResourceType)
			}
			if data.ResourceName != tt.wantResource {
				t.Errorf("ResourceName = %v, want %v", data.ResourceName, tt.wantResource)
			}
			if data.IsSuccessful != tt.wantSuccessful {
				t.Errorf("IsSuccessful = %v, want %v", data.IsSuccessful, tt.wantSuccessful)
			}
			if data.ProjectID != tt.wantProjectID {
				t.Errorf("ProjectID = %v, want %v", data.ProjectID, tt.wantProjectID)
			}
			if len(tt.wantDesc) > 0 && data.OperationDescription != tt.wantDesc {
				t.Errorf("OperationDescription = %q, want %q", data.OperationDescription, tt.wantDesc)
			}
		})
	}
}

func TestParsePreResolved(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantType string
	}{
		{"g:developers", "developers", "g"},
		{"u:admin", "admin", "u"},
		{"plain", "plain", ""},
		{"", "", ""},
		{"g:name:with:colons", "name:with:colons", "g"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, typ := parsePreResolved(tt.input)
			if name != tt.wantName || typ != tt.wantType {
				t.Errorf("parsePreResolved(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, typ, tt.wantName, tt.wantType)
			}
		})
	}
}
