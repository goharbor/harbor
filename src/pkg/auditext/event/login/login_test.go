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

package login

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

func Test_login_Resolve(t *testing.T) {
	type args struct {
		ce    *commonevent.Metadata
		event *event.Event
	}
	tests := []struct {
		name                     string
		l                        *loginResolver
		args                     args
		wantErr                  bool
		wantUsername             string
		wantOperation            string
		wantOperationDescription string
		wantIsSuccessful         bool
	}{

		{"test normal", &loginResolver{}, args{
			ce: &commonevent.Metadata{
				Username:      "test",
				RequestURL:    "/c/login",
				RequestMethod: "POST",
				Payload:       "principal=test&password=123456",
				ResponseCode:  200,
			}, event: &event.Event{}}, false, "test", "login", "login", true},
		{"test fail", &loginResolver{}, args{
			ce: &commonevent.Metadata{
				Username:      "test",
				RequestURL:    "/c/login",
				RequestMethod: "POST",
				Payload:       "principal=test&password=123456",
				ResponseCode:  401,
			}, event: &event.Event{}}, false, "test", "login", "login", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &loginResolver{}
			if err := l.Resolve(tt.args.ce, tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("resolver.Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.event.Data.(*model.CommonEvent).Operator != tt.wantUsername {
				t.Errorf("resolver.Resolve() got = %v, want %v", tt.args.event.Data.(*model.CommonEvent).Operator, tt.wantUsername)
			}
			if tt.args.event.Data.(*model.CommonEvent).Operation != tt.wantOperation {
				t.Errorf("resolver.Resolve() got = %v, want %v", tt.args.event.Data.(*model.CommonEvent).Operation, tt.wantOperation)
			}
			if tt.args.event.Data.(*model.CommonEvent).OperationDescription != tt.wantOperationDescription {
				t.Errorf("resolver.Resolve() got = %v, want %v", tt.args.event.Data.(*model.CommonEvent).OperationDescription, tt.wantOperationDescription)
			}
			if tt.args.event.Data.(*model.CommonEvent).IsSuccessful != tt.wantIsSuccessful {
				t.Errorf("resolver.Resolve() got = %v, want %v", tt.args.event.Data.(*model.CommonEvent).IsSuccessful, tt.wantIsSuccessful)
			}
		})
	}
}

func Test_login_PreCheck(t *testing.T) {
	type args struct {
		ctx    context.Context
		url    string
		method string
	}
	tests := []struct {
		name             string
		e                *loginResolver
		args             args
		wantMatched      bool
		wantResourceName string
	}{
		{"test normal", &loginResolver{}, args{context.Background(), "/c/login", "POST"}, true, ""},
		{"test fail method", &loginResolver{}, args{context.Background(), "/c/login", "PUT"}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &loginResolver{}
			got, got1 := e.PreCheck(tt.args.ctx, tt.args.url, tt.args.method)
			if got != tt.wantMatched {
				t.Errorf("resolver.PreCheck() got = %v, want %v", got, tt.wantMatched)
			}
			if got1 != tt.wantResourceName {
				t.Errorf("resolver.PreCheck() got1 = %v, want %v", got1, tt.wantResourceName)
			}
		})
	}
}
