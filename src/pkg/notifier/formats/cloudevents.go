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

package formats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

var (
	// cloudEventsFormatter is the global single formatter for CloudEvents.
	cloudEventsFormatter Formatter = &CloudEvents{}
)

func init() {
	registerFormats(CloudEventsFormat, cloudEventsFormatter)
}

const (
	// CloudEventsFormat is the type for CloudEvents format.
	CloudEventsFormat = "CloudEvents"
	// extRequestID is the key for the request id in the CloudEvents.
	extRequestID = "requestid"
	// extOperator is the key for the operator in the CloudEvents.
	extOperator = "operator"
)

var (
	// eventTypeMapping defines the mapping of harbor event type and CloudEvents type.
	eventTypeMapping = map[string]string{
		event.TopicDeleteArtifact:    eventType("artifact.deleted"),
		event.TopicPullArtifact:      eventType("artifact.pulled"),
		event.TopicPushArtifact:      eventType("artifact.pushed"),
		event.TopicQuotaExceed:       eventType("quota.exceeded"),
		event.TopicQuotaWarning:      eventType("quota.warned"),
		event.TopicReplication:       eventType("replication.status.changed"),
		event.TopicScanningFailed:    eventType("scan.failed"),
		event.TopicScanningCompleted: eventType("scan.completed"),
		event.TopicScanningStopped:   eventType("scan.stopped"),
		event.TopicTagRetention:      eventType("tag_retention.finished"),
	}
)

// eventType returns the constructed event type.
func eventType(t string) string {
	// defines the prefix for event type, organization name or FQDN or more extended possibility,
	// use harbor by default.
	prefix := "harbor"
	return fmt.Sprintf("%s.%s", prefix, t)
}

// CloudEvents is the instance for the CloudEvents format.
type CloudEvents struct{}

// Format implements the interface Formatter.
/*
{
   "specversion":"1.0",
   "requestid": "2eedfab8-61d3-4f3c-8ec3-8f82d1ec4c84",
   "id":"4b2f89a6-548d-4c12-9993-a1f5790b97d2",
   "source":"/projects/1/webhook/policies/3",
   "type":"harbor.artifact.pulled",
   "datacontenttype":"application/json",
   "time":"2023-03-06T06:08:43Z",
   "data":{
      "resources":[
         {
            "digest":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
            "tag":"sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c",
            "resource_url":"harbor.dev/library/busybox@sha256:dde8e930c7b6a490f728e66292bc9bce42efc9bbb5278bae40e4f30f6e00fe8c"
         }
      ],
      "repository":{
         "date_created":1677053165,
         "name":"busybox",
         "namespace":"library",
         "repo_full_name":"library/busybox",
         "repo_type":"public"
      }
   },
   "operator":"robot$library+scanner-Trivy-51fe4548-bbe5-11ed-9217-0242ac14000d"
}
*/
func (ce *CloudEvents) Format(ctx context.Context, he *model.HookEvent) (http.Header, []byte, error) {
	if he == nil {
		return nil, nil, errors.Errorf("HookEvent should not be nil")
	}

	eventType, ok := eventTypeMapping[he.EventType]
	if !ok {
		return nil, nil, errors.Errorf("unknown event type: %s", he.EventType)
	}

	event := cloudevents.NewEvent()
	// the cloudEvents id is uuid, but we insert the request id as extension which can be used to trace the event.
	event.SetID(uuid.NewString())
	event.SetExtension(extRequestID, lib.GetXRequestID(ctx))
	event.SetSource(source(he.ProjectID, he.PolicyID))
	event.SetType(eventType)
	event.SetTime(time.Unix(he.Payload.OccurAt, 0))
	event.SetExtension(extOperator, he.Payload.Operator)

	if err := event.SetData(cloudevents.ApplicationJSON, he.Payload.EventData); err != nil {
		return nil, nil, errors.Wrap(err, "error to set data in CloudEvents")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error to marshal CloudEvents")
	}

	header := http.Header{
		"Content-Type": []string{cloudevents.ApplicationCloudEventsJSON},
	}
	return header, data, nil
}

// source builds the source for CloudEvents.
func source(projectID, policyID int64) string {
	return fmt.Sprintf("/projects/%d/webhook/policies/%d", projectID, policyID)
}
