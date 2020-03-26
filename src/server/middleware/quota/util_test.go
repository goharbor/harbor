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

package quota

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/testing/api/project"
	"github.com/stretchr/testify/mock"
)

func Test_projectReferenceObject(t *testing.T) {
	ctl := &project.Controller{}
	ctl.On("GetByName", mock.AnythingOfType(""), "library").Return(&models.Project{ProjectID: 1}, nil)
	ctl.On("GetByName", mock.AnythingOfType(""), "demo").Return(nil, fmt.Errorf("not found"))

	originalProjectController := projectController
	defer func() {
		projectController = originalProjectController
	}()

	projectController = ctl

	req := func(path string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, path, nil)

		return r.WithContext(context.TODO())
	}

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{"/api/v2.0/projects/library", args{req("/api/v2.0/projects/library")}, "project", "1", false},
		{"/api/v2.0/projects/library/repositories", args{req("/api/v2.0/projects/library/repositories")}, "project", "1", false},
		{"/api/v2.0/projects/demo", args{req("/api/v2.0/projects/demo")}, "", "", true},
		{"/api/v2.0/library", args{req("/api/v2.0/library")}, "", "", true},
		{"/v2/library/photon/manifests/2.0", args{req("/v2/library/photon/manifests/2.0")}, "project", "1", false},
		{"/v2", args{req("/v2")}, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := projectReferenceObject(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("projectReferenceObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("projectReferenceObject() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("projectReferenceObject() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
