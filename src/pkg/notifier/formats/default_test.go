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

package formats

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

func TestDefault_Format(t *testing.T) {
	type args struct {
		he *model.HookEvent
	}
	tests := []struct {
		name    string
		d       *Default
		args    args
		want    http.Header
		want1   []byte
		wantErr bool
	}{
		{
			name:    "invalid case",
			d:       &Default{},
			args:    args{he: nil},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "normal case",
			d:    &Default{},
			args: args{he: &model.HookEvent{
				Payload: &model.Payload{
					Type:     "PULL_ARTIFACT",
					OccurAt:  1678082303,
					Operator: "admin",
					EventData: &model.EventData{
						Resources: []*model.Resource{
							{Digest: "sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
								Tag:         "sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
								ResourceURL: "harbor.dev/library/busybox@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
							},
						},
						Repository: &model.Repository{
							DateCreated:  1677053165,
							Name:         "busybox",
							Namespace:    "library",
							RepoFullName: "library/busybox",
							RepoType:     "public",
						},
					},
				},
			}},
			want:    http.Header{"Content-Type": []string{"application/json"}},
			want1:   []byte(`{"type":"PULL_ARTIFACT","occur_at":1678082303,"operator":"admin","event_data":{"resources":[{"digest":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c","tag":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c","resource_url":"harbor.dev/library/busybox@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c"}],"repository":{"date_created":1677053165,"name":"busybox","namespace":"library","repo_full_name":"library/busybox","repo_type":"public"}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Default{}
			got, got1, err := d.Format(context.TODO(), tt.args.he)
			if (err != nil) != tt.wantErr {
				t.Errorf("Default.Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Default.Format() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Default.Format() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
