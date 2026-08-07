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

package project

import (
	"context"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/controller/event/metadata/commonevent"
	"github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	mockproject "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProjectEventResolver_PreCheck(t *testing.T) {
	c := &resolver{}

	tests := []struct {
		name   string
		url    string
		method string
		want   bool
	}{
		{"test PUT", "/api/v2.0/projects/1", "PUT", true},
		{"test POST", "/api/v2.0/projects/1/metadatas/", "POST", true},
		{"test DELETE project metadata", "/api/v2.0/projects/1/metadatas/proxy_cache_filter_pattern", "DELETE", true},
		{"test DELETE project", "/api/v2.0/projects/1", "DELETE", false},
		{"test GET", "/api/v2.0/projects/1", "GET", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := c.PreCheck(context.Background(), tt.url, tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetProjectNameOrID(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"normal project url", "/api/v2.0/projects/myproj", "myproj"},
		{"normal project url with query", "/api/v2.0/projects/myproj?x_is_resource_name=true", "myproj"},
		{"metadatas url", "/api/v2.0/projects/myproj/metadatas/", "myproj"},
		{"metadatas url with query", "/api/v2.0/projects/myproj/metadatas/?x_is_resource_name=true", "myproj"},
		{"metadata key url", "/api/v2.0/projects/myproj/metadatas/proxy_cache_filter_pattern", "myproj"},
		{"metadata key url with query", "/api/v2.0/projects/myproj/metadatas/proxy_cache_filter_pattern?x_is_resource_name=true", "myproj"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getProjectNameOrID(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProjectEventResolver_Resolve(t *testing.T) {
	mockCtrl := &mockproject.Controller{}
	oldCtrl := project.Ctl
	project.Ctl = mockCtrl
	defer func() { project.Ctl = oldCtrl }()

	proj := &proModels.Project{
		ProjectID: 123,
		Name:      "test-proj",
	}

	mockCtrl.On("Get", mock.Anything, int64(123)).Return(proj, nil)
	mockCtrl.On("Get", mock.Anything, "test-proj").Return(proj, nil)

	r := &resolver{}

	t.Run("Resolve project update by ID", func(t *testing.T) {
		ce := &commonevent.Metadata{
			Ctx:           context.Background(),
			Username:      "admin",
			RequestURL:    "/api/v2.0/projects/123",
			RequestMethod: "PUT",
			ResponseCode:  http.StatusOK,
		}
		evt := &event.Event{}
		err := r.Resolve(ce, evt)
		assert.NoError(t, err)

		data, ok := evt.Data.(*model.CommonEvent)
		assert.True(t, ok)
		assert.Equal(t, "update", data.Operation)
		assert.Equal(t, "admin", data.Operator)
		assert.Equal(t, "project", data.ResourceType)
		assert.Equal(t, "test-proj", data.ResourceName)
		assert.Equal(t, int64(123), data.ProjectID)
		assert.Equal(t, "update project: test-proj", data.OperationDescription)
		assert.True(t, data.IsSuccessful)
	})

	t.Run("Resolve project update by Name", func(t *testing.T) {
		ce := &commonevent.Metadata{
			Ctx:           context.Background(),
			Username:      "admin",
			RequestURL:    "/api/v2.0/projects/test-proj/metadatas/",
			RequestMethod: "POST",
			ResponseCode:  http.StatusOK,
		}
		evt := &event.Event{}
		err := r.Resolve(ce, evt)
		assert.NoError(t, err)

		data, ok := evt.Data.(*model.CommonEvent)
		assert.True(t, ok)
		assert.Equal(t, "update", data.Operation)
		assert.Equal(t, "admin", data.Operator)
		assert.Equal(t, "project", data.ResourceType)
		assert.Equal(t, "test-proj", data.ResourceName)
		assert.Equal(t, int64(123), data.ProjectID)
		assert.Equal(t, "update project: test-proj", data.OperationDescription)
		assert.True(t, data.IsSuccessful)
	})
}
