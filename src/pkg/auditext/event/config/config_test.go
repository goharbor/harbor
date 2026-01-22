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

package config

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

func TestConfigureEventResolver_Resolve(t *testing.T) {
	type fields struct {
		SensitiveAttributes []string
	}
	type args struct {
		ce  *commonevent.Metadata
		evt *event.Event
	}
	tests := []struct {
		name                     string
		fields                   fields
		args                     args
		wantErr                  bool
		wantOperation            string
		wantResourceType         string
		wantOperationDescription string
	}{
		{"test normal", fields{[]string{}}, args{
			ce: &commonevent.Metadata{
				Username:      "test",
				RequestURL:    "/api/v2.0/configurations",
				RequestMethod: "PUT",
			}, evt: &event.Event{}}, false, "update", "configuration", "update configuration"},
		{"test error", fields{[]string{}}, args{ce: &commonevent.Metadata{
			Username:      "test",
			RequestURL:    "/api/v2.0/configurations",
			RequestMethod: "POST",
		}, evt: &event.Event{}}, false, "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &resolver{
				SensitiveAttributes: tt.fields.SensitiveAttributes,
			}
			if err := c.Resolve(tt.args.ce, tt.args.evt); (err != nil) != tt.wantErr {
				t.Errorf("ConfigureEventResolver.Resolve() error = %v, wantErr %v", err, tt.wantErr)
				if tt.args.evt.Data != nil {
					data := tt.args.evt.Data.(*model.CommonEvent)
					if data.Operation != tt.wantOperation {
						t.Errorf("ConfigureEventResolver.Resolve() Operation = %v, want %v", data.Operation, tt.wantOperation)
					}
					if data.ResourceType != tt.wantResourceType {
						t.Errorf("ConfigureEventResolver.Resolve() ResourceType = %v, want %v", data.ResourceType, tt.wantResourceType)
					}
					if data.OperationDescription != tt.wantOperationDescription {
						t.Errorf("ConfigureEventResolver.Resolve() OperationDescription = %v, want %v", data.OperationDescription, tt.wantOperationDescription)
					}
				}

			}
		})
	}
}

func TestConfigureEventResolver_PreCheck(t *testing.T) {
	type fields struct {
		SensitiveAttributes []string
	}
	type args struct {
		ctx    context.Context
		url    string
		method string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      bool
		wantTopic string
	}{
		{"test normal", fields{[]string{}}, args{context.Background(), "/api/v2.0/configurations", "PUT"}, true, ""},
		{"test error", fields{[]string{}}, args{context.Background(), "/api/v2.0/configurations", "POST"}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &resolver{
				SensitiveAttributes: tt.fields.SensitiveAttributes,
			}
			got, got1 := c.PreCheck(tt.args.ctx, tt.args.url, tt.args.method)
			if got != tt.want {
				t.Errorf("ConfigureEventResolver.PreCheck() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantTopic {
				t.Errorf("ConfigureEventResolver.PreCheck() got1 = %v, want %v", got1, tt.wantTopic)
			}
		})
	}
}
