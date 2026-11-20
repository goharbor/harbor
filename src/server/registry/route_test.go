// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/project"
	_ "github.com/goharbor/harbor/src/pkg/auditext/event/config"
	projectmodels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/server/middleware/artifactinfo"
	logmiddleware "github.com/goharbor/harbor/src/server/middleware/log"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
)

type routeObservedBody struct {
	remaining int64
	bytesRead int64
}

func (r *routeObservedBody) Read(buffer []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	count := int64(len(buffer))
	if count > r.remaining {
		count = r.remaining
	}
	for i := int64(0); i < count; i++ {
		buffer[i] = 'x'
	}
	r.remaining -= count
	r.bytesRead += count
	return int(count), nil
}

func (*routeObservedBody) Close() error {
	return nil
}

func TestManifestRouteLimitsBodyBeforeReaders(t *testing.T) {
	originalApp := web.BeeApp
	web.BeeApp = web.NewHttpSever()
	defer func() {
		web.BeeApp = originalApp
	}()

	originalProjectController := project.Ctl
	projectController := &projecttesting.Controller{}
	mock.OnAnything(projectController, "Get").Return(
		&projectmodels.Project{ProjectID: 1, Name: "project"},
		nil,
	)
	project.Ctl = projectController
	defer func() {
		project.Ctl = originalProjectController
	}()

	RegisterRoutes()

	securityContext := &securitytesting.Context{}
	mock.OnAnything(securityContext, "Can").Return(true)
	securityContext.On("IsAuthenticated").Return(true)

	handler := artifactinfo.Middleware()(logmiddleware.Middleware()(web.BeeApp.Handlers))
	for _, unknownLength := range []bool{false, true} {
		name := "known length"
		if unknownLength {
			name = "unknown length"
		}
		t.Run(name, func(t *testing.T) {
			bodySize := int64(2 * common.MaxManifestBodySize)
			body := &routeObservedBody{remaining: bodySize}
			request := httptest.NewRequest(
				http.MethodPut,
				"/v2/project/repo/manifests/latest?x=/api/v2.0/configurations",
				body,
			)
			request.ContentLength = bodySize
			wantBytesRead := int64(0)
			if unknownLength {
				request.ContentLength = -1
				wantBytesRead = common.MaxManifestBodySize + 1
			}
			request = request.WithContext(security.NewContext(context.Background(), securityContext))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusRequestEntityTooLarge {
				t.Errorf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
			}
			if body.bytesRead != wantBytesRead {
				t.Errorf("body bytes read = %d, want %d", body.bytesRead, wantBytesRead)
			}
			assertRouteRequestEntityTooLarge(t, response)
		})
	}
}

func assertRouteRequestEntityTooLarge(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var payload struct {
		Errors []struct {
			Code string `json:"code"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if len(payload.Errors) != 1 || payload.Errors[0].Code != "REQUEST_ENTITY_TOO_LARGE" {
		t.Errorf("error response = %s, want REQUEST_ENTITY_TOO_LARGE", response.Body.String())
	}
}
