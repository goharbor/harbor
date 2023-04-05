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
	"encoding/json"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/stretchr/testify/assert"
)

func TestCloudEvents_Format(t *testing.T) {
	ce := &CloudEvents{}
	// invalid case
	{
		header, data, err := ce.Format(nil, nil)
		assert.Error(t, err)
		assert.Nil(t, header)
		assert.Nil(t, data)
	}
	// normal case
	{
		he := &model.HookEvent{
			ProjectID: 1,
			PolicyID:  3,
			EventType: "PULL_ARTIFACT",
			Payload: &model.Payload{
				Type:     "PULL_ARTIFACT",
				OccurAt:  1678082923,
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
		}

		ctx := context.TODO()
		requestID := "mock-request-id"
		header, data, err := ce.Format(lib.WithXRequestID(ctx, requestID), he)
		assert.NoError(t, err)
		assert.Equal(t, http.Header{"Content-Type": []string{"application/cloudevents+json"}}, header)
		// validate data format
		event := cloudevents.NewEvent()
		err = json.Unmarshal(data, &event)
		assert.NoError(t, err)
		assert.Equal(t, "1.0", event.SpecVersion())
		assert.Equal(t, requestID, event.Extensions()["requestid"])
		assert.Equal(t, "/projects/1/webhook/policies/3", event.Source())
		assert.Equal(t, "harbor.artifact.pulled", event.Type())
		assert.Equal(t, "application/json", event.DataContentType())
		assert.Equal(t, "2023-03-06T06:08:43Z", event.Time().Format(time.RFC3339))
		assert.Equal(t, "admin", event.Extensions()["operator"])
		assert.Equal(t, `{"resources":[{"digest":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c","tag":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c","resource_url":"harbor.dev/library/busybox@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c"}],"repository":{"date_created":1677053165,"name":"busybox","namespace":"library","repo_full_name":"library/busybox","repo_type":"public"}}`, string(event.Data()))
	}
}
